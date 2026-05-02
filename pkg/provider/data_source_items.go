package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// Ensure the implementation satisfies the expected interfaces
type itemsDataSource struct {
	client *apiclient.Client
}

// itemsDataSourceModel maps the data source schema data
type itemsDataSourceModel struct {
	ItemID              types.String `tfsdk:"item_id"`
	Path                types.String `tfsdk:"path"`
	FieldNames          types.List   `tfsdk:"field_names"`
	ExistingVersionOnly types.Bool   `tfsdk:"existing_version_only"`
	Required            types.Bool   `tfsdk:"required"`
	Items               types.List   `tfsdk:"items"`
}

// childItemModel represents a single child item in the response
type childItemModel struct {
	ItemID       types.String `tfsdk:"item_id"`
	Path         types.String `tfsdk:"path"`
	Name         types.String `tfsdk:"name"`
	DisplayName  types.String `tfsdk:"display_name"`
	TemplateID   types.String `tfsdk:"template_id"`
	TemplateName types.String `tfsdk:"template_name"`
	Fields       types.Map    `tfsdk:"fields"`
}

// NewItemsDataSource creates a new data source
func NewItemsDataSource() datasource.DataSource {
	return &itemsDataSource{}
}

// Metadata returns the data source type name
func (d *itemsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_items"
}

// Configure sets the client for the data source
func (d *itemsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Schema defines the schema for the data source
func (d *itemsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the children of a Sitecore item by ID or path.",
		Attributes: map[string]schema.Attribute{
			"item_id": schema.StringAttribute{
				Description: "The ID of the parent item to retrieve children for.",
				Optional:    true,
			},
			"path": schema.StringAttribute{
				Description: "The path of the parent item to retrieve children for.",
				Optional:    true,
			},
			"field_names": schema.ListAttribute{
				Description: "The names of the fields to retrieve for each child item.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"existing_version_only": schema.BoolAttribute{
				Description: "Whether to only retrieve existing versions of items.",
				Optional:    true,
			},
			"required": schema.BoolAttribute{
				Description: "Whether the parent item must exist (default: true).",
				Optional:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "The child items of the parent item.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"item_id": schema.StringAttribute{
							Description: "The ID of the child item.",
							Computed:    true,
						},
						"path": schema.StringAttribute{
							Description: "The path of the child item.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the child item.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the child item.",
							Computed:    true,
						},
						"template_id": schema.StringAttribute{
							Description: "The template ID of the child item.",
							Computed:    true,
						},
						"template_name": schema.StringAttribute{
							Description: "The template name of the child item.",
							Computed:    true,
						},
						"fields": schema.MapAttribute{
							Description: "The fields of the child item.",
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Read implements the data source read operation
func (d *itemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state itemsDataSourceModel

	// Get current state
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either item_id or path is specified
	if (state.ItemID.IsNull() || len(state.ItemID.ValueString()) == 0) && (state.Path.IsNull() || len(state.Path.ValueString()) == 0) {
		resp.Diagnostics.AddError(
			"Missing Required Parameter",
			"Either item_id or path must be specified to retrieve children.",
		)
		return
	}

	// Validate that only one of item_id or path is specified
	if (!state.ItemID.IsNull() && len(state.ItemID.ValueString()) > 0) && (!state.Path.IsNull() && len(state.Path.ValueString()) > 0) {
		resp.Diagnostics.AddError(
			"Conflicting Parameters",
			"Only one of item_id or path should be specified, not both.",
		)
		return
	}

	// Get field names if specified, or use default fields
	var fieldNames []string
	if !state.FieldNames.IsNull() && len(state.FieldNames.Elements()) > 0 {
		diags := state.FieldNames.ElementsAs(ctx, &fieldNames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Use default field names when none are specified
		fieldNames = []string{"title", "text", "description", "tenantName"}
	}

	// Note: Field mapping is handled internally by the API client

	// Get existingVersionOnly parameter if specified
	var existingVersionOnly *bool
	if !state.ExistingVersionOnly.IsNull() {
		boolValue := state.ExistingVersionOnly.ValueBool()
		existingVersionOnly = &boolValue
	}

	// Get required parameter if specified, default to true for backward compatibility
	isRequired := true
	if !state.Required.IsNull() {
		isRequired = state.Required.ValueBool()
	}

	// Retrieve parent item to get children
	var parentItem *apiclient.Item
	var err error

	if !state.ItemID.IsNull() && len(state.ItemID.ValueString()) > 0 {
		parentItem, err = d.client.GetItemByID(state.ItemID.ValueString(), fieldNames, existingVersionOnly)
		if err != nil {
			// Check if the error is about item not found and if the item is not required
			if !isRequired && (err.Error() == fmt.Sprintf("item with ID '%s' not found", state.ItemID.ValueString()) || err.Error() == "item with ID '"+state.ItemID.ValueString()+"' not found") {
				// Item not found but not required, continue with empty items list
				parentItem = nil
			} else {
				resp.Diagnostics.AddError(
					"Unable to Read Sitecore Item",
					"Unable to retrieve parent item by ID "+state.ItemID.ValueString()+": "+err.Error(),
				)
				return
			}
		}
	} else {
		parentItem, err = d.client.GetItemByPath(state.Path.ValueString(), fieldNames, existingVersionOnly)
		if err != nil {
			// Check if the error is about item not found and if the item is not required
			if !isRequired && (err.Error() == fmt.Sprintf("item with path '%s' not found", state.Path.ValueString()) || err.Error() == "item with path '"+state.Path.ValueString()+"' not found") {
				// Item not found but not required, continue with empty items list
				parentItem = nil
			} else {
				resp.Diagnostics.AddError(
					"Unable to Read Sitecore Item",
					"Unable to retrieve parent item by path "+state.Path.ValueString()+": "+err.Error(),
				)
				return
			}
		}
	}

	// Map children to response
	var itemModels []childItemModel
	if parentItem != nil && len(parentItem.Children) > 0 {
		// Create field name mapping for children
		fieldMapping := make(map[string]string)
		for i, fieldName := range fieldNames {
			fieldAlias := fmt.Sprintf("field%d", i+1)
			fieldMapping[fieldAlias] = fieldName
		}

		for _, child := range parentItem.Children {
			// Convert fields to map[string]attr.Value for Terraform
			fieldsMap := make(map[string]attr.Value)

			// First, ensure all requested fields are present (even if null)
			for _, fieldName := range fieldNames {
				fieldsMap[fieldName] = types.StringNull()
			}

			// Then populate with actual values from API response
			for key, value := range child.Fields {
				if value == nil {
					// Field is null, already set to null above
					continue
				} else if strValue, ok := value.(string); ok {
					// Field has a string value
					if originalName, exists := fieldMapping[key]; exists {
						fieldsMap[originalName] = types.StringValue(strValue)
					}
				} else {
					// Fallback for other types (shouldn't happen after transformation)
					if originalName, exists := fieldMapping[key]; exists {
						fieldsMap[originalName] = types.StringValue(fmt.Sprintf("%v", value))
					}
				}
			}

			childItemModel := childItemModel{
				ItemID:       types.StringValue(child.ItemID),
				Path:         types.StringValue(child.Path),
				Name:         types.StringValue(child.Name),
				DisplayName:  types.StringValue(child.DisplayName),
				TemplateID:   types.StringValue(child.TemplateID),
				TemplateName: types.StringValue(child.TemplateName),
				Fields:       types.MapValueMust(types.StringType, fieldsMap),
			}
			itemModels = append(itemModels, childItemModel)
		}
	}

	// Convert item models to list value
	itemElements := make([]attr.Value, len(itemModels))
	for i, model := range itemModels {
		itemObject, diags := types.ObjectValue(
			map[string]attr.Type{
				"item_id":       types.StringType,
				"path":          types.StringType,
				"name":          types.StringType,
				"display_name":  types.StringType,
				"template_id":   types.StringType,
				"template_name": types.StringType,
				"fields":        types.MapType{ElemType: types.StringType},
			},
			map[string]attr.Value{
				"item_id":       model.ItemID,
				"path":          model.Path,
				"name":          model.Name,
				"display_name":  model.DisplayName,
				"template_id":   model.TemplateID,
				"template_name": model.TemplateName,
				"fields":        model.Fields,
			},
		)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		itemElements[i] = itemObject
	}

	itemsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"item_id":       types.StringType,
				"path":          types.StringType,
				"name":          types.StringType,
				"display_name":  types.StringType,
				"template_id":   types.StringType,
				"template_name": types.StringType,
				"fields":        types.MapType{ElemType: types.StringType},
			},
		},
		itemElements,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set the items in the state
	state.Items = itemsList

	// Set the state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
