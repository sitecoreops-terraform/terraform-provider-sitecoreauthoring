package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &itemResource{}
	_ resource.ResourceWithConfigure   = &itemResource{}
	_ resource.ResourceWithImportState = &itemResource{}
)

// NewItemResource is a helper function to simplify the provider implementation
func NewItemResource() resource.Resource {
	return &itemResource{}
}

// itemResource is the resource implementation
type itemResource struct {
	client *apiclient.Client
}

// itemResourceModel maps the resource schema data to a Go type
type itemResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ItemID     types.String `tfsdk:"item_id"`
	Path       types.String `tfsdk:"path"`
	Name       types.String `tfsdk:"name"`
	ParentID   types.String `tfsdk:"parent_id"`
	TemplateID types.String `tfsdk:"template_id"`
	Language   types.String `tfsdk:"language"`
	Fields     types.Map    `tfsdk:"fields"`
	Database   types.String `tfsdk:"database"`
}

// convertFieldsMap converts a types.Map with string elements to map[string]interface{}
func convertFieldsMap(ctx context.Context, fieldsMap types.Map) (map[string]interface{}, error) {
	if fieldsMap.IsNull() || fieldsMap.IsUnknown() {
		return nil, nil
	}

	// Get the elements from the map
	elements := fieldsMap.Elements()
	if len(elements) == 0 {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{})

	// Iterate through each element in the map
	for key, value := range elements {
		// Convert the attr.Value to string
		if !value.IsNull() && !value.IsUnknown() {
			strValue := value.(types.String).ValueString()
			result[key] = strValue
		} else {
			result[key] = nil
		}
	}

	return result, nil
}

// Metadata returns the resource type name
func (r *itemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

// Schema defines the schema for the resource
func (r *itemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sitecore item using the Authoring API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the item resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"item_id": schema.StringAttribute{
				Description: "The ID of the item.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Description: "The path of the item. This is computed based on the item's location and name in Sitecore.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					pathComputedOnNameChangeModifier{},
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the item.",
				Required:    true,
			},
			"parent_id": schema.StringAttribute{
				Description: "The ID of the parent item.",
				Required:    true,
			},
			"template_id": schema.StringAttribute{
				Description: "The template ID for the item.",
				Required:    true,
			},
			"language": schema.StringAttribute{
				Description: "The language of the item.",
				Required:    true,
			},
			"fields": schema.MapAttribute{
				Description: "The fields of the item as key-value pairs.",
				ElementType: types.StringType,
				Optional:    true,
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
func (r *itemResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*apiclient.Client)
}

// Create creates the resource and sets the initial Terraform state
func (r *itemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan itemResourceModel

	// Get the plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert fields from Terraform format to map[string]interface{}
	fieldsMap := make(map[string]interface{})
	if !plan.Fields.IsNull() && len(plan.Fields.Elements()) > 0 {
		var err error
		fieldsMap, err = convertFieldsMap(ctx, plan.Fields)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Convert Fields",
				"Unable to convert fields: "+err.Error(),
			)
			return
		}
	}

	// Create the item using the API client
	item, err := r.client.CreateItem(
		plan.Name.ValueString(),
		plan.TemplateID.ValueString(),
		plan.ParentID.ValueString(),
		plan.Language.ValueString(),
		fieldsMap,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Sitecore Item",
			"Unable to create item: "+err.Error(),
		)
		return
	}

	// Map the response to the resource model
	state := convertToResourceModel(item, plan)

	// Set the state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data
func (r *itemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state itemResourceModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use path if available, otherwise use item_id
	var item *apiclient.Item
	var err error

	if !state.Path.IsNull() && len(state.Path.ValueString()) > 0 {
		item, err = r.client.GetItemByPathWithFields(state.Path.ValueString(), nil)
	} else if !state.ItemID.IsNull() && len(state.ItemID.ValueString()) > 0 {
		item, err = r.client.GetItemByIDWithFields(state.ItemID.ValueString(), nil)
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Identifier",
			"Either path or item_id must be available to read the item.",
		)
		return
	}

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

	// Update the state with the latest data
	updatedState := convertToResourceModel(item, state)

	// Set the state
	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state
func (r *itemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan itemResourceModel
	var state itemResourceModel

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

	// Check if the name is being changed
	nameChanged := !state.Name.Equal(plan.Name)

	database := "master"
	if !plan.Database.IsNull() && len(plan.Database.ValueString()) > 0 {
		database = plan.Database.ValueString()
	}

	var item *apiclient.Item
	var err error

	// Convert fields from Terraform format to map[string]interface{}
	fieldsMap := make(map[string]interface{})
	if !plan.Fields.IsNull() && len(plan.Fields.Elements()) > 0 {
		var err error
		fieldsMap, err = convertFieldsMap(ctx, plan.Fields)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Convert Fields",
				"Unable to convert fields: "+err.Error(),
			)
			return
		}
	}

	if nameChanged {
		// If name is being changed, use RenameItem method first
		item, err = r.client.RenameItem(
			state.ItemID.ValueString(),
			plan.Name.ValueString(),
			database,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Rename Sitecore Item",
				"Unable to rename item: "+err.Error(),
			)
			return
		}

		// If fields are also being changed, update them on the renamed item
		if len(fieldsMap) > 0 {
			item, err = r.client.UpdateItem(
				state.ItemID.ValueString(),
				plan.Language.ValueString(),
				fieldsMap,
				database,
				item.Path, // Use the new path from the renamed item
			)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Sitecore Item After Rename",
					"Unable to update item fields after rename: "+err.Error(),
				)
				return
			}
		}
	} else {
		// If name is not being changed, use UpdateItem method for field updates
		item, err = r.client.UpdateItem(
			state.ItemID.ValueString(),
			plan.Language.ValueString(),
			fieldsMap,
			database,
			state.Path.ValueString(),
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Sitecore Item",
				"Unable to update item: "+err.Error(),
			)
			return
		}
	}

	// Map the response to the resource model
	updatedState := convertToResourceModel(item, plan)

	// Set the state
	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource
func (r *itemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state itemResourceModel

	// Get current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the item using the API client
	success, err := r.client.DeleteItem(state.Path.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Sitecore Item",
			"Unable to delete item: "+err.Error(),
		)
		return
	}

	if !success {
		resp.Diagnostics.AddError(
			"Failed to Delete Sitecore Item",
			"Item deletion was not successful.",
		)
		return
	}

	// Remove the resource from state
	resp.State.RemoveResource(ctx)
}

// ImportState imports an existing resource into Terraform state
func (r *itemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// convertToResourceModel converts an API client Item to the resource model
func convertToResourceModel(item *apiclient.Item, plan itemResourceModel) itemResourceModel {
	// Convert fields to map[string]attr.Value for Terraform
	var fieldsMap map[string]attr.Value

	// Process item fields if they exist
	if len(item.Fields) > 0 {
		fieldsMap = make(map[string]attr.Value)
		for key, value := range item.Fields {
			if value == nil {
				fieldsMap[key] = types.StringNull()
			} else if strValue, ok := value.(string); ok {
				fieldsMap[key] = types.StringValue(strValue)
			} else {
				fieldsMap[key] = types.StringValue(fmt.Sprintf("%v", value))
			}
		}
	} else if !plan.Fields.IsNull() {
		// If item has no fields but plan fields are not null, create empty map
		fieldsMap = make(map[string]attr.Value)
	} else {
		// If both item has no fields and plan fields are null, keep null
		fieldsMap = nil
	}

	// Create the resource model
	model := itemResourceModel{
		ID:         types.StringValue(item.ItemID),
		ItemID:     types.StringValue(item.ItemID),
		Path:       types.StringValue(item.Path),
		Name:       types.StringValue(item.Name),
		ParentID:   plan.ParentID,               // Keep parent_id from plan
		TemplateID: plan.TemplateID,             // Keep template_id from plan
		Language:   plan.Language,               // Keep language from plan
		Database:   types.StringValue("master"), // Default to master
	}

	// Handle fields separately to maintain null vs empty map consistency
	if fieldsMap == nil {
		model.Fields = types.MapNull(types.StringType)
	} else {
		model.Fields = types.MapValueMust(types.StringType, fieldsMap)
	}

	// If the plan had specific values for database, preserve them
	if !plan.Database.IsNull() && len(plan.Database.ValueString()) > 0 {
		model.Database = plan.Database
	}

	return model
}
