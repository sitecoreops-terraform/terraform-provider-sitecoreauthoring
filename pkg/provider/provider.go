//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name sitecoreauthoring

// Package provider contains the Sitecore Authoring Terraform provider implementation
package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &sitecoreProvider{}
)

// New is a helper function to simplify provider server and testing implementation
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &sitecoreProvider{
			version: version,
		}
	}
}

// sitecoreProvider is the provider implementation
type sitecoreProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// sitecoreProviderModel maps provider schema data to a Go type
type sitecoreProviderModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Host         types.String `tfsdk:"host"`
	UseCLI       types.Bool   `tfsdk:"use_cli"`
	CLIEndpoint  types.String `tfsdk:"cli_endpoint"`
}

// Metadata returns the provider type name
func (p *sitecoreProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sitecoreauthoring"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *sitecoreProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Sitecore Authoring API",
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "The client ID for Sitecore Authoring API authentication. Can also be specified with env var SITECORE_AUTHORING_CLIENT_ID.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The client secret for Sitecore Authoring API authentication. Can also be specified with env var SITECORE_AUTHORING_CLIENT_SECRET.",
				Optional:    true,
				Sensitive:   true,
			},
			"host": schema.StringAttribute{
				Description: "The host URL for Sitecore Authoring API. Can also be specified with env var SITECORE_AUTHORING_HOST.",
				Optional:    true,
			},
			"use_cli": schema.BoolAttribute{
				Description: "Use Sitecore CLI authentication (uses .sitecore/user.json file)",
				Optional:    true,
			},
			"cli_endpoint": schema.StringAttribute{
				Description: "The endpoint name to use when using CLI authentication. Can also be specified with env var SITECORE_AUTHORING_CLI_ENDPOINT.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a Sitecore Authoring API client for data sources and resources
func (p *sitecoreProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config sitecoreProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var client *apiclient.Client
	var err error

	// Handle environment variables
	// Check if CLI authentication is requested
	useCLI := os.Getenv("SITECORE_AUTHORING_USE_CLI") == "1" || os.Getenv("SITECORE_AUTHORING_USE_CLI") == "true"
	if !config.UseCLI.IsNull() {
		useCLI = config.UseCLI.ValueBool()
	}

	// Get CLI endpoint name
	cliEndpoint := os.Getenv("SITECORE_AUTHORING_CLI_ENDPOINT")
	if !config.CLIEndpoint.IsNull() && len(config.CLIEndpoint.ValueString()) > 0 {
		cliEndpoint = config.CLIEndpoint.ValueString()
	}

	// Use traditional client_id/client_secret authentication
	clientID := os.Getenv("SITECORE_AUTHORING_CLIENT_ID")
	clientSecret := os.Getenv("SITECORE_AUTHORING_CLIENT_SECRET")
	host := os.Getenv("SITECORE_AUTHORING_HOST")

	// Override with configuration values if provided
	if !config.ClientID.IsNull() && len(config.ClientID.ValueString()) > 0 {
		clientID = config.ClientID.ValueString()
	}
	if !config.ClientSecret.IsNull() && len(config.ClientSecret.ValueString()) > 0 {
		clientSecret = config.ClientSecret.ValueString()
	}
	if !config.Host.IsNull() && len(config.Host.ValueString()) > 0 {
		host = config.Host.ValueString()
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value
	if config.ClientID.IsUnknown() || config.ClientSecret.IsUnknown() || config.Host.IsUnknown() || config.UseCLI.IsUnknown() || config.CLIEndpoint.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Sitecore Authoring API Configuration",
			"Cannot use unknown values for Sitecore API configuration",
		)
		return
	}

	// Validate that host is provided when not using CLI authentication
	if !useCLI && len(host) == 0 {
		resp.Diagnostics.AddError(
			"Missing Sitecore Authoring API Host",
			"Host must be provided when using client_id and client_secret authentication. Set host parameter or SITECORE_AUTHORING_HOST environment variable.",
		)
		return
	}

	if useCLI {
		client, err = apiclient.NewClientFromCLIWithEndpoint("", cliEndpoint)
		if err != nil {
			resp.Diagnostics.AddError(
				"Sitecore Authoring CLI Authentication Failed",
				"Unable to authenticate using Sitecore CLI: "+err.Error(),
			)
			return
		}
	} else {
		client, err = apiclient.NewClientWithAllConfig(host, "", clientID, clientSecret, "", &http.Client{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Sitecore Authoring API Authentication Failed",
				"Unable to authenticate using Client Id and Client Secret: "+err.Error(),
			)
			return
		}
	}

	// Authenticate the client
	err = client.Authenticate()
	if err != nil {
		resp.Diagnostics.AddError(
			"Sitecore Authoring API Client Authentication Failed",
			"Unable to authenticate Sitecore API client: "+err.Error(),
		)
		return
	}

	// Make the Sitecore API client available during data source and resource
	// type Configure methods
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider
func (p *sitecoreProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSitesDataSource,
		NewItemDataSource,
		NewItemsDataSource,
	}
}

// Resources defines the resources implemented in the provider
func (p *sitecoreProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewItemResource,
		NewItemFieldResource,
	}
}
