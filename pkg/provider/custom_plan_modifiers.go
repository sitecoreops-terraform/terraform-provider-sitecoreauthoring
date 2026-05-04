package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// pathComputedOnNameChangeModifier ensures path is computed when item name changes.
// This modifier addresses the Sitecore-specific behavior where renaming an item
// changes its path, which would otherwise cause Terraform to report inconsistent results.
type pathComputedOnNameChangeModifier struct{}

func (m pathComputedOnNameChangeModifier) Description(ctx context.Context) string {
	return "Ensures path is computed when item name changes"
}

func (m pathComputedOnNameChangeModifier) MarkdownDescription(ctx context.Context) string {
	return "The path will be computed when the item name changes, as renaming an item changes its path in Sitecore. This prevents 'inconsistent result' errors during rename operations."
}

func (m pathComputedOnNameChangeModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If this is a create operation, no need to modify
	if req.State.Raw.IsNull() {
		return
	}

	// Get the current state and planned values
	var state, plan itemResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if name is changing
	nameChanged := !state.Name.Equal(plan.Name)

	// If name is changing, mark path as unknown so it will be computed
	if nameChanged {
		resp.PlanValue = types.StringUnknown()
	}
	// Otherwise, keep the existing behavior
}
