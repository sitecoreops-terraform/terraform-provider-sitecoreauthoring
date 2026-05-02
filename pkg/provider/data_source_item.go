package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &itemDataSource{}
	_ datasource.DataSourceWithConfigure = &itemDataSource{}
)

// NewItemDataSource is a helper function to simplify the provider implementation
func NewItemDataSource() datasource.DataSource {
	return &itemDataSource{}
}

// itemDataSource is the data source implementation
type itemDataSource struct {
	client *apiclient.Client
}

// itemDataSourceModel maps the data source schema data to a Go type
type itemDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ItemID              types.String `tfsdk:"item_id"`
	Path                types.String `tfsdk:"path"`
	FieldNames          types.Set    `tfsdk:"field_names"`
	ExistingVersionOnly types.Bool   `tfsdk:"existing_version_only"`
	Required            types.Bool   `tfsdk:"required"`
	Item                *itemModel   `tfsdk:"item"`
}

type itemModel struct {
	ItemID       types.String `tfsdk:"item_id"`
	Path         types.String `tfsdk:"path"`
	Name         types.String `tfsdk:"name"`
	DisplayName  types.String `tfsdk:"display_name"`
	TemplateID   types.String `tfsdk:"template_id"`
	TemplateName types.String `tfsdk:"template_name"`
	Fields       types.Map    `tfsdk:"fields"`
}

// Metadata returns the data source type name
func (d *itemDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_item"
}

// Schema defines the schema for the data source
func (d *itemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about a Sitecore item by ID or path from the Authoring API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
				Computed:    true,
			},
			"item_id": schema.StringAttribute{
				Description: "The ID of the item to retrieve. Either item_id or path must be specified.",
				Optional:    true,
			},
			"path": schema.StringAttribute{
				Description: "The path of the item to retrieve. Either item_id or path must be specified.",
				Optional:    true,
			},
			"field_names": schema.SetAttribute{
				Description: "Set of field names to retrieve for the item and its children. If not specified, no fields will be retrieved.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"existing_version_only": schema.BoolAttribute{
				Description: "If true, only returns items that have an existing version. If not specified, returns items regardless of version status.",
				Optional:    true,
			},
			"required": schema.BoolAttribute{
				Description: "If true, the data source will fail if the item doesn't exist. If false or not specified, the data source will return null values if the item doesn't exist.",
				Optional:    true,
			},
			"item": schema.SingleNestedAttribute{
				Description: "The retrieved item.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"item_id": schema.StringAttribute{
						Description: "The ID of the item.",
						Computed:    true,
					},
					"path": schema.StringAttribute{
						Description: "The path of the item.",
						Computed:    true,
					},
					"name": schema.StringAttribute{
						Description: "The name of the item.",
						Computed:    true,
					},
					"display_name": schema.StringAttribute{
						Description: "The display name of the item.",
						Computed:    true,
					},
					"template_id": schema.StringAttribute{
						Description: "The template ID of the item.",
						Computed:    true,
					},
					"template_name": schema.StringAttribute{
						Description: "The template name of the item.",
						Computed:    true,
					},
					"fields": schema.MapAttribute{
						Description: "The fields of the item.",
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *itemDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*apiclient.Client)
}

// Read refreshes the Terraform state with the latest data
func (d *itemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state itemDataSourceModel

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
			"Either item_id or path must be specified to retrieve an item.",
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

	// Get field names if specified
	var fieldNames []string
	if !state.FieldNames.IsNull() && len(state.FieldNames.Elements()) > 0 {
		diags := state.FieldNames.ElementsAs(ctx, &fieldNames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

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

	var item *apiclient.Item
	var err error

	// Retrieve item by ID or path
	if !state.ItemID.IsNull() && len(state.ItemID.ValueString()) > 0 {
		item, err = d.client.GetItemByID(state.ItemID.ValueString(), fieldNames, existingVersionOnly)
		if err != nil {
			// Check if the error is about item not found and if the item is not required
			if !isRequired && (err.Error() == fmt.Sprintf("item with ID '%s' not found", state.ItemID.ValueString()) || err.Error() == "item with ID '"+state.ItemID.ValueString()+"' not found") {
				// Item not found but not required, continue with null values
				item = nil
			} else {
				resp.Diagnostics.AddError(
					"Unable to Read Sitecore Item",
					"Unable to retrieve item by ID "+state.ItemID.ValueString()+": "+err.Error(),
				)
				return
			}
		}
	} else {
		item, err = d.client.GetItemByPath(state.Path.ValueString(), fieldNames, existingVersionOnly)
		if err != nil {
			// Check if the error is about item not found and if the item is not required
			if !isRequired && (err.Error() == fmt.Sprintf("item with path '%s' not found", state.Path.ValueString()) || err.Error() == "item with path '"+state.Path.ValueString()+"' not found") {
				// Item not found but not required, continue with null values
				item = nil
			} else {
				resp.Diagnostics.AddError(
					"Unable to Read Sitecore Item",
					"Unable to retrieve item by path "+state.Path.ValueString()+": "+err.Error(),
				)
				return
			}
		}
	}

	// Map response to state
	if item != nil {
		// Convert fields to map[string]attr.Value for Terraform
		fieldsMap := make(map[string]attr.Value)

		// Ensure all requested fields are present, even if null
		for _, fieldName := range fieldNames {
			fieldsMap[fieldName] = types.StringNull()
		}

		// Then populate with actual values from API response
		for key, value := range item.Fields {
			if value == nil {
				// Field is null, already set to null above
				continue
			} else if strValue, ok := value.(string); ok {
				// Field has a string value
				// Check if this is an alias and map to original field name
				if len(key) > 5 && key[:5] == "field" {
					// This is an alias like "field1", find which field it corresponds to
					if i, err := strconv.Atoi(key[5:]); err == nil && i > 0 && i <= len(fieldNames) {
						originalFieldName := fieldNames[i-1]
						fieldsMap[originalFieldName] = types.StringValue(strValue)
					}
				} else {
					// Not an alias, use as-is
					fieldsMap[key] = types.StringValue(strValue)
				}
			} else {
				// Fallback for other types (shouldn't happen after transformation)
				// Check if this is an alias and map to original field name
				if len(key) > 5 && key[:5] == "field" {
					// This is an alias like "field1", find which field it corresponds to
					if i, err := strconv.Atoi(key[5:]); err == nil && i > 0 && i <= len(fieldNames) {
						originalFieldName := fieldNames[i-1]
						fieldsMap[originalFieldName] = types.StringValue(fmt.Sprintf("%v", value))
					}
				} else {
					// Not an alias, use as-is
					fieldsMap[key] = types.StringValue(fmt.Sprintf("%v", value))
				}
			}
		}

		itemModel := itemModel{
			ItemID:       types.StringValue(item.ItemID),
			Path:         types.StringValue(item.Path),
			Name:         types.StringValue(item.Name),
			DisplayName:  types.StringValue(item.DisplayName),
			TemplateID:   types.StringValue(item.TemplateID),
			TemplateName: types.StringValue(item.TemplateName),
			Fields:       types.MapValueMust(types.StringType, fieldsMap),
		}

		state.Item = &itemModel
	}

	// Set the ID to the item ID or path
	if !state.ItemID.IsNull() && len(state.ItemID.ValueString()) > 0 {
		state.ID = types.StringValue(state.ItemID.ValueString())
	} else {
		state.ID = types.StringValue(state.Path.ValueString())
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
