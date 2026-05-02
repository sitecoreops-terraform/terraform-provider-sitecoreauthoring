package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ datasource.DataSource              = &sitesDataSource{}
	_ datasource.DataSourceWithConfigure = &sitesDataSource{}
)

// NewSitesDataSource is a helper function to simplify the provider implementation
func NewSitesDataSource() datasource.DataSource {
	return &sitesDataSource{}
}

// sitesDataSource is the data source implementation
type sitesDataSource struct {
	client *apiclient.Client
}

// sitesDataSourceModel maps the data source schema data to a Go type
type sitesDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Sites      []siteModel  `tfsdk:"sites"`
	SiteNames  types.Set    `tfsdk:"site_names"`
	SearchName types.String `tfsdk:"search_name"`
	RootPath   types.String `tfsdk:"root_path"`
}

type siteModel struct {
	Name            types.String `tfsdk:"name"`
	HostName        types.String `tfsdk:"host_name"`
	TargetHostName  types.String `tfsdk:"target_host_name"`
	ContentLanguage types.String `tfsdk:"content_language"`
	Language        types.String `tfsdk:"language"`
	Domain          types.String `tfsdk:"domain"`
	RootPath        types.String `tfsdk:"root_path"`
	StartPath       types.String `tfsdk:"start_path"`
	BrowserTitle    types.String `tfsdk:"browser_title"`
	CacheHtml       types.Bool   `tfsdk:"cache_html"`
	CacheMedia      types.Bool   `tfsdk:"cache_media"`
	EnablePreview   types.Bool   `tfsdk:"enable_preview"`
	RootItemID      types.String `tfsdk:"root_item_id"`
	StartItemID     types.String `tfsdk:"start_item_id"`
	StartItemPath   types.String `tfsdk:"start_item_path"`
}

// Metadata returns the data source type name
func (d *sitesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sites"
}

// Schema defines the schema for the data source
func (d *sitesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about Sitecore sites from the Authoring API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
				Computed:    true,
			},
			"sites": schema.ListNestedAttribute{
				Description: "List of sites.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the site.",
							Computed:    true,
						},
						"host_name": schema.StringAttribute{
							Description: "The host name of the site.",
							Computed:    true,
						},
						"target_host_name": schema.StringAttribute{
							Description: "The target host name of the site.",
							Computed:    true,
						},
						"content_language": schema.StringAttribute{
							Description: "The content language of the site.",
							Computed:    true,
						},
						"language": schema.StringAttribute{
							Description: "The language of the site.",
							Computed:    true,
						},
						"domain": schema.StringAttribute{
							Description: "The domain of the site.",
							Computed:    true,
						},
						"root_path": schema.StringAttribute{
							Description: "The root path of the site.",
							Computed:    true,
						},
						"start_path": schema.StringAttribute{
							Description: "The start path of the site.",
							Computed:    true,
						},
						"browser_title": schema.StringAttribute{
							Description: "The browser title of the site.",
							Computed:    true,
						},
						"cache_html": schema.BoolAttribute{
							Description: "Whether HTML caching is enabled for the site.",
							Computed:    true,
						},
						"cache_media": schema.BoolAttribute{
							Description: "Whether media caching is enabled for the site.",
							Computed:    true,
						},
						"enable_preview": schema.BoolAttribute{
							Description: "Whether preview is enabled for the site.",
							Computed:    true,
						},
						"root_item_id": schema.StringAttribute{
							Description: "The root item ID of the site.",
							Computed:    true,
						},
						"start_item_id": schema.StringAttribute{
							Description: "The start item ID of the site.",
							Computed:    true,
						},
						"start_item_path": schema.StringAttribute{
							Description: "The start item path of the site.",
							Computed:    true,
						},
					},
				},
			},
			"site_names": schema.SetAttribute{
				Description: "Set of site names for filtering. If specified, only sites with these names will be returned.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"search_name": schema.StringAttribute{
				Description: "Name of a specific site to search for. If specified, only the site with this exact name will be returned.",
				Optional:    true,
			},
			"root_path": schema.StringAttribute{
				Description: "Root path to filter sites by. If specified, only sites with root paths starting with this value will be returned (supports prefix matching).",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *sitesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*apiclient.Client)
}

// Read refreshes the Terraform state with the latest data
func (d *sitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sitesDataSourceModel

	// Get current state
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all sites from API
	sites, err := d.client.GetSites()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Sitecore Sites",
			"Unable to retrieve sites from Sitecore Authoring API: "+err.Error(),
		)
		return
	}

	// Filter sites based on criteria
	var filteredSites []apiclient.Site

	if !state.SearchName.IsNull() && len(state.SearchName.ValueString()) > 0 {
		// Search for specific site by name
		searchName := state.SearchName.ValueString()
		site, err := d.client.GetSiteByName(searchName)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Find Site",
				"Unable to find site with name "+searchName+": "+err.Error(),
			)
			return
		}
		if site != nil {
			filteredSites = append(filteredSites, *site)
		}
	} else if !state.RootPath.IsNull() && len(state.RootPath.ValueString()) > 0 {
		// Filter by root path (supports wildcard/prefix matching)
		rootPath := state.RootPath.ValueString()
		for _, site := range sites {
			if strings.HasPrefix(site.RootPath, rootPath) {
				filteredSites = append(filteredSites, site)
			}
		}
	} else if !state.SiteNames.IsNull() && len(state.SiteNames.Elements()) > 0 {
		// Filter by site names set
		var siteNames []string
		diags := state.SiteNames.ElementsAs(ctx, &siteNames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		nameMap := make(map[string]bool)
		for _, name := range siteNames {
			nameMap[name] = true
		}

		for _, site := range sites {
			if nameMap[site.Name] {
				filteredSites = append(filteredSites, site)
			}
		}
	} else {
		// Return all sites
		filteredSites = sites
	}

	// Map response to state
	siteModels := make([]siteModel, 0, len(filteredSites))
	for _, site := range filteredSites {
		siteModel := siteModel{
			Name:            types.StringValue(site.Name),
			HostName:        types.StringValue(site.HostName),
			TargetHostName:  types.StringValue(site.TargetHostName),
			ContentLanguage: types.StringValue(site.ContentLanguage.Name),
			Language:        types.StringValue(site.Language),
			Domain:          types.StringValue(site.Domain),
			RootPath:        types.StringValue(site.RootPath),
			StartPath:       types.StringValue(site.StartPath),
			BrowserTitle:    types.StringValue(site.BrowserTitle),
			CacheHtml:       types.BoolValue(site.CacheHtml),
			CacheMedia:      types.BoolValue(site.CacheMedia),
			EnablePreview:   types.BoolValue(site.EnablePreview),
			RootItemID:      types.StringValue(site.RootItem.ItemID),
			StartItemID:     types.StringValue(site.StartItem.ItemID),
			StartItemPath:   types.StringValue(site.StartItem.Path),
		}
		siteModels = append(siteModels, siteModel)
	}

	// Set the ID to a placeholder value
	state.ID = types.StringValue("sites")
	state.Sites = siteModels

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
