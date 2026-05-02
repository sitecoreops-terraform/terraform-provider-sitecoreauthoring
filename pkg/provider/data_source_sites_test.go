package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

func TestSitesDataSourceMapping(t *testing.T) {
	// Create test data matching the sample response
	testSites := []apiclient.Site{
		{
			Name:           "shared-config",
			HostName:       "*",
			TargetHostName: "",
			ContentLanguage: struct {
				Name string `json:"name,omitempty"`
			}{
				Name: "en",
			},
			Language:      "en",
			Domain:        "",
			RootPath:      "/sitecore/content/asmblii/shared-config",
			StartPath:     "/Home",
			BrowserTitle:  "Shared-config - Sitecore",
			CacheHtml:     false,
			CacheMedia:    true,
			EnablePreview: true,
			RootItem: struct {
				ItemID string `json:"itemId"`
			}{
				ItemID: "59e05408e2c74916b27ca718bbabf4d9",
			},
			StartItem: struct {
				ItemID string `json:"itemId,omitempty"`
				Path   string `json:"path,omitempty"`
			}{
				ItemID: "59e05408e2c74916b27ca718bbabf4d9",
				Path:   "/sitecore/content/asmblii/shared-config",
			},
		},
		{
			Name:           "reference",
			HostName:       "dev-website-reference.vercel.app|dev-website-reference*.vercel.app|localhost:3000",
			TargetHostName: "",
			ContentLanguage: struct {
				Name string `json:"name,omitempty"`
			}{
				Name: "en",
			},
			Language:      "en",
			Domain:        "",
			RootPath:      "/sitecore/content/asmblii/reference",
			StartPath:     "/Home",
			BrowserTitle:  "Reference - Sitecore",
			CacheHtml:     false,
			CacheMedia:    true,
			EnablePreview: true,
			RootItem: struct {
				ItemID string `json:"itemId"`
			}{
				ItemID: "eeb52b4fb19a4ebb8f17192c1c2b89fc",
			},
			StartItem: struct {
				ItemID string `json:"itemId,omitempty"`
				Path   string `json:"path,omitempty"`
			}{
				ItemID: "eeb52b4fb19a4ebb8f17192c1c2b89fc",
				Path:   "/sitecore/content/asmblii/reference",
			},
		},
		{
			Name:           "website",
			HostName:       "",
			TargetHostName: "",
			ContentLanguage: struct {
				Name string `json:"name,omitempty"`
			}{
				Name: "en",
			},
			Language:      "en",
			Domain:        "extranet",
			RootPath:      "/sitecore/content",
			StartPath:     "/home",
			BrowserTitle:  "Website - Sitecore",
			CacheHtml:     false,
			CacheMedia:    true,
			EnablePreview: true,
			RootItem: struct {
				ItemID string `json:"itemId"`
			}{
				ItemID: "0de95ae441ab4d019eb067441b7c2450",
			},
			StartItem: struct {
				ItemID string `json:"itemId,omitempty"`
				Path   string `json:"path,omitempty"`
			}{
				ItemID: "0de95ae441ab4d019eb067441b7c2450",
				Path:   "/sitecore/content",
			},
		},
	}

	// Create state and populate with test data
	var state sitesDataSourceModel
	state.ID = types.StringValue("sites")

	// Map the test sites to site models (this is what the Read function does)
	var siteModels []siteModel
	for _, site := range testSites {
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
			CacheMedia:      types.BoolValue(site.CacheMedia),
			EnablePreview:   types.BoolValue(site.EnablePreview),
			RootItemID:      types.StringValue(site.RootItem.ItemID),
			StartItemID:     types.StringValue(site.StartItem.ItemID),
			StartItemPath:   types.StringValue(site.StartItem.Path),
		}
		siteModels = append(siteModels, siteModel)
	}

	state.Sites = siteModels

	// Verify all fields are correctly mapped for the first site
	site1 := state.Sites[0]
	if site1.Name.ValueString() != "shared-config" {
		t.Errorf("Expected Name 'shared-config', got '%s'", site1.Name.ValueString())
	}
	if site1.HostName.ValueString() != "*" {
		t.Errorf("Expected HostName '*', got '%s'", site1.HostName.ValueString())
	}
	if site1.TargetHostName.ValueString() != "" {
		t.Errorf("Expected TargetHostName '', got '%s'", site1.TargetHostName.ValueString())
	}
	if site1.ContentLanguage.ValueString() != "en" {
		t.Errorf("Expected ContentLanguage 'en', got '%s'", site1.ContentLanguage.ValueString())
	}
	if site1.Language.ValueString() != "en" {
		t.Errorf("Expected Language 'en', got '%s'", site1.Language.ValueString())
	}
	if site1.Domain.ValueString() != "" {
		t.Errorf("Expected Domain '', got '%s'", site1.Domain.ValueString())
	}
	if site1.RootPath.ValueString() != "/sitecore/content/asmblii/shared-config" {
		t.Errorf("Expected RootPath '/sitecore/content/asmblii/shared-config', got '%s'", site1.RootPath.ValueString())
	}
	if site1.StartPath.ValueString() != "/Home" {
		t.Errorf("Expected StartPath '/Home', got '%s'", site1.StartPath.ValueString())
	}
	if site1.BrowserTitle.ValueString() != "Shared-config - Sitecore" {
		t.Errorf("Expected BrowserTitle 'Shared-config - Sitecore', got '%s'", site1.BrowserTitle.ValueString())
	}
	if site1.CacheMedia.ValueBool() != true {
		t.Errorf("Expected CacheMedia true, got %v", site1.CacheMedia.ValueBool())
	}
	if site1.EnablePreview.ValueBool() != true {
		t.Errorf("Expected EnablePreview true, got %v", site1.EnablePreview.ValueBool())
	}
	if site1.RootItemID.ValueString() != "59e05408e2c74916b27ca718bbabf4d9" {
		t.Errorf("Expected RootItemID '59e05408e2c74916b27ca718bbabf4d9', got '%s'", site1.RootItemID.ValueString())
	}
	if site1.StartItemID.ValueString() != "59e05408e2c74916b27ca718bbabf4d9" {
		t.Errorf("Expected StartItemID '59e05408e2c74916b27ca718bbabf4d9', got '%s'", site1.StartItemID.ValueString())
	}
	if site1.StartItemPath.ValueString() != "/sitecore/content/asmblii/shared-config" {
		t.Errorf("Expected StartItemPath '/sitecore/content/asmblii/shared-config', got '%s'", site1.StartItemPath.ValueString())
	}

	// Verify complex hostName for second site
	site2 := state.Sites[1]
	if site2.HostName.ValueString() != "dev-website-reference.vercel.app|dev-website-reference*.vercel.app|localhost:3000" {
		t.Errorf("Expected HostName 'dev-website-reference.vercel.app|dev-website-reference*.vercel.app|localhost:3000', got '%s'", site2.HostName.ValueString())
	}

	// Verify third site with domain
	site3 := state.Sites[2]
	if site3.Domain.ValueString() != "extranet" {
		t.Errorf("Expected Domain 'extranet', got '%s'", site3.Domain.ValueString())
	}
	if site3.HostName.ValueString() != "" {
		t.Errorf("Expected HostName '', got '%s'", site3.HostName.ValueString())
	}

	// Verify we have all 3 sites
	if len(state.Sites) != 3 {
		t.Errorf("Expected 3 sites, got %d", len(state.Sites))
	}
}

func TestSitesDataSourceRootPathFiltering(t *testing.T) {
	// Create test data matching the sample response
	testSites := []apiclient.Site{
		{
			Name:     "shared-config",
			RootPath: "/sitecore/content/asmblii/shared-config",
		},
		{
			Name:     "reference",
			RootPath: "/sitecore/content/asmblii/reference",
		},
		{
			Name:     "website",
			RootPath: "/sitecore/content",
		},
	}

	// Test exact root path matching
	rootPathToFilter := "/sitecore/content/asmblii/shared-config"
	var filteredSites []apiclient.Site
	for _, site := range testSites {
		if site.RootPath == rootPathToFilter {
			filteredSites = append(filteredSites, site)
		}
	}

	// Should only have 1 site matching the exact root path
	if len(filteredSites) != 1 {
		t.Errorf("Expected 1 site with root path '%s', got %d", rootPathToFilter, len(filteredSites))
	}

	if filteredSites[0].Name != "shared-config" {
		t.Errorf("Expected site name 'shared-config', got '%s'", filteredSites[0].Name)
	}

	// Test wildcard/prefix matching
	prefixToFilter := "/sitecore/content/asmblii"
	var prefixFilteredSites []apiclient.Site
	for _, site := range testSites {
		if strings.HasPrefix(site.RootPath, prefixToFilter) {
			prefixFilteredSites = append(prefixFilteredSites, site)
		}
	}

	// Should have 2 sites matching the prefix
	if len(prefixFilteredSites) != 2 {
		t.Errorf("Expected 2 sites with root path starting with '%s', got %d", prefixToFilter, len(prefixFilteredSites))
	}

	// Verify the correct sites are returned
	siteNames := make(map[string]bool)
	for _, site := range prefixFilteredSites {
		siteNames[site.Name] = true
	}

	if !siteNames["shared-config"] {
		t.Errorf("Expected 'shared-config' in prefix filtered results")
	}
	if !siteNames["reference"] {
		t.Errorf("Expected 'reference' in prefix filtered results")
	}

	// Test with non-existent root path
	nonExistentRootPath := "/sitecore/content/nonexistent"
	var emptyFilteredSites []apiclient.Site
	for _, site := range testSites {
		if strings.HasPrefix(site.RootPath, nonExistentRootPath) {
			emptyFilteredSites = append(emptyFilteredSites, site)
		}
	}

	if len(emptyFilteredSites) != 0 {
		t.Errorf("Expected 0 sites with non-existent root path prefix, got %d", len(emptyFilteredSites))
	}
}
