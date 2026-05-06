package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource              = &itemFieldResource{}
	_ resource.ResourceWithConfigure = &itemFieldResource{}
)

// NewItemFieldResource is a helper function to simplify the provider implementation
func NewItemFieldResource() resource.Resource {
	return &itemFieldResource{}
}

// itemFieldResource is the resource implementation
type itemFieldResource struct {
	client *apiclient.Client
}

// itemFieldResourceModel maps the resource schema data to a Go type
type itemFieldResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ItemID     types.String `tfsdk:"item_id"`
	Language   types.String `tfsdk:"language"`
	FieldName  types.String `tfsdk:"field_name"`
	FieldValue types.String `tfsdk:"field_value"`
	Database   types.String `tfsdk:"database"`
}

// Metadata returns the resource type name
func (r *itemFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item_field"
}

// Schema defines the schema for the resource
func (r *itemFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single field value on an existing Sitecore item using the Authoring API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the item field resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"item_id": schema.StringAttribute{
				Description: "The ID of the item.",
				Required:    true,
			},
			"language": schema.StringAttribute{
				Description: "The language of the item.",
				Required:    true,
			},
			"field_name": schema.StringAttribute{
				Description: "The name of the field to update.",
				Required:    true,
			},
			"field_value": schema.StringAttribute{
				Description: "The value to set for the field.",
				Required:    true,
			},
			"database": schema.StringAttribute{
				Description: "The database where the item is stored (default: master).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *itemFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*apiclient.Client)
}

// Create creates the resource and sets the initial Terraform state
func (r *itemFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan itemFieldResourceModel

	// Get the plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the fields map for UpdateItem
	fieldsMap := map[string]interface{}{
		plan.FieldName.ValueString(): plan.FieldValue.ValueString(),
	}

	database := "master"
	if !plan.Database.IsNull() && len(plan.Database.ValueString()) > 0 {
		database = plan.Database.ValueString()
	}

	// Get the current item to get the path
	item, err := r.client.GetItemByID(plan.ItemID.ValueString(), []string{}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Sitecore Item",
			"Unable to get item: "+err.Error(),
		)
		return
	}

	// Update the item field using the API client
	updatedItem, err := r.client.UpdateItem(
		plan.ItemID.ValueString(),
		plan.Language.ValueString(),
		fieldsMap,
		database,
		item.Path,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Sitecore Item Field",
			"Unable to update item field: "+err.Error(),
		)
		return
	}

	// Map the response to the resource model
	state := convertToItemFieldResourceModel(updatedItem, plan)

	// Ensure the field value is preserved from the plan if the API response doesn't contain it
	// This handles cases where UpdateItem doesn't return the updated field values
	if state.FieldValue.ValueString() == "" && plan.FieldValue.ValueString() != "" {
		state.FieldValue = plan.FieldValue
	}

	// Set the state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data
func (r *itemFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state itemFieldResourceModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the item to read the current field value using GetItemByIDWithFields
	// This ensures we get all fields including the one we're managing
	item, err := r.client.GetItemByIDWithFields(state.ItemID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Sitecore Item",
			"Unable to retrieve item: "+err.Error(),
		)
		return
	}

	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if the field exists and get its value
	fieldValue, exists := item.Fields[state.FieldName.ValueString()]
	if !exists {
		// If the field doesn't exist, it might have been cleared, so we should remove the resource
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state with the current field value
	updatedState := state
	if fieldValue != nil {
		if strValue, ok := fieldValue.(string); ok {
			updatedState.FieldValue = types.StringValue(strValue)
		} else {
			updatedState.FieldValue = types.StringValue(fmt.Sprintf("%v", fieldValue))
		}
	} else {
		// If field value is nil, treat it as empty string to maintain consistency
		updatedState.FieldValue = types.StringValue("")
	}

	// Set the state
	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state
func (r *itemFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan itemFieldResourceModel
	var state itemFieldResourceModel

	// Get the current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the plan
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	database := "master"
	if !plan.Database.IsNull() && len(plan.Database.ValueString()) > 0 {
		database = plan.Database.ValueString()
	}

	// Prepare the fields map for UpdateItem
	fieldsMap := map[string]interface{}{
		plan.FieldName.ValueString(): plan.FieldValue.ValueString(),
	}

	// Get the current item to get the path
	item, err := r.client.GetItemByID(plan.ItemID.ValueString(), []string{}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Sitecore Item",
			"Unable to get item: "+err.Error(),
		)
		return
	}

	// Update the item field using the API client
	updatedItem, err := r.client.UpdateItem(
		plan.ItemID.ValueString(),
		plan.Language.ValueString(),
		fieldsMap,
		database,
		item.Path,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Sitecore Item Field",
			"Unable to update item field: "+err.Error(),
		)
		return
	}

	// Map the response to the resource model
	updatedState := convertToItemFieldResourceModel(updatedItem, plan)

	// Ensure the field value is preserved from the plan if the API response doesn't contain it
	// This handles cases where UpdateItem doesn't return the updated field values
	if updatedState.FieldValue.ValueString() == "" && plan.FieldValue.ValueString() != "" {
		updatedState.FieldValue = plan.FieldValue
	}

	// Set the state
	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource
func (r *itemFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state itemFieldResourceModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the resource from state
	resp.State.RemoveResource(ctx)
}

// convertToItemFieldResourceModel converts an API client Item to the item field resource model
func convertToItemFieldResourceModel(item *apiclient.Item, plan itemFieldResourceModel) itemFieldResourceModel {
	// Get the field value from the item
	fieldValue := ""
	if item.Fields != nil {
		if val, exists := item.Fields[plan.FieldName.ValueString()]; exists {
			if val != nil {
				if strValue, ok := val.(string); ok {
					fieldValue = strValue
				} else {
					fieldValue = fmt.Sprintf("%v", val)
				}
			}
		}
	}

	// Create the resource model
	model := itemFieldResourceModel{
		ID:         types.StringValue(fmt.Sprintf("%s-%s-%s", item.ItemID, plan.Language.ValueString(), plan.FieldName.ValueString())),
		ItemID:     types.StringValue(item.ItemID),
		Language:   types.StringValue(plan.Language.ValueString()),
		FieldName:  types.StringValue(plan.FieldName.ValueString()),
		FieldValue: types.StringValue(fieldValue),
		Database:   types.StringValue("master"),
	}

	// If the plan had specific values for database, preserve them
	if !plan.Database.IsNull() && len(plan.Database.ValueString()) > 0 {
		model.Database = plan.Database
	}

	return model
}
