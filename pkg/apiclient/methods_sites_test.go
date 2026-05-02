package apiclient

import (
	"testing"
)

func TestParseSitesResponse(t *testing.T) {
	// Sample response data matching the provided JSON
	responseData := map[string]interface{}{
		"sites": []interface{}{
			map[string]interface{}{
				"name":            "shared-config",
				"hostName":        "*",
				"targetHostName":  "",
				"contentLanguage": map[string]interface{}{"name": "en"},
				"language":        "en",
				"domain":          nil,
				"rootPath":        "/sitecore/content/asmblii/shared-config",
				"startPath":       "/Home",
				"browserTitle":    "Shared-config - Sitecore",
				"cacheHtml":       false,
				"cacheMedia":      true,
				"enablePreview":   true,
				"rootItem":        map[string]interface{}{"itemId": "59e05408e2c74916b27ca718bbabf4d9"},
				"startItem":       map[string]interface{}{"itemId": "59e05408e2c74916b27ca718bbabf4d9", "path": "/sitecore/content/asmblii/shared-config"},
			},
			map[string]interface{}{
				"name":            "reference",
				"hostName":        "dev-website-reference.vercel.app|dev-website-reference*.vercel.app|localhost:3000",
				"targetHostName":  "",
				"contentLanguage": map[string]interface{}{"name": "en"},
				"language":        "en",
				"domain":          nil,
				"rootPath":        "/sitecore/content/asmblii/reference",
				"startPath":       "/Home",
				"browserTitle":    "Reference - Sitecore",
				"cacheHtml":       false,
				"cacheMedia":      true,
				"enablePreview":   true,
				"rootItem":        map[string]interface{}{"itemId": "eeb52b4fb19a4ebb8f17192c1c2b89fc"},
				"startItem":       map[string]interface{}{"itemId": "eeb52b4fb19a4ebb8f17192c1c2b89fc", "path": "/sitecore/content/asmblii/reference"},
			},
			map[string]interface{}{
				"name":            "website",
				"hostName":        "",
				"targetHostName":  "",
				"contentLanguage": map[string]interface{}{"name": "en"},
				"language":        "en",
				"domain":          "extranet",
				"rootPath":        "/sitecore/content",
				"startPath":       "/home",
				"browserTitle":    "Website - Sitecore",
				"cacheHtml":       false,
				"cacheMedia":      true,
				"enablePreview":   true,
				"rootItem":        map[string]interface{}{"itemId": "0de95ae441ab4d019eb067441b7c2450"},
				"startItem":       map[string]interface{}{"itemId": "0de95ae441ab4d019eb067441b7c2450", "path": "/sitecore/content"},
			},
		},
	}

	var sitesResponse SitesResponse
	err := parseGraphQLResponse(responseData, &sitesResponse)
	if err != nil {
		t.Fatalf("Failed to parse sites response: %v", err)
	}

	// Verify we got the expected number of sites
	if len(sitesResponse.Sites) != 3 {
		t.Fatalf("Expected 3 sites, got %d", len(sitesResponse.Sites))
	}

	// Test first site (shared-config)
	site1 := sitesResponse.Sites[0]
	if site1.Name != "shared-config" {
		t.Errorf("Expected site name 'shared-config', got '%s'", site1.Name)
	}
	if site1.HostName != "*" {
		t.Errorf("Expected hostName '*', got '%s'", site1.HostName)
	}
	if site1.TargetHostName != "" {
		t.Errorf("Expected targetHostName '', got '%s'", site1.TargetHostName)
	}
	if site1.ContentLanguage.Name != "en" {
		t.Errorf("Expected contentLanguage.name 'en', got '%s'", site1.ContentLanguage.Name)
	}
	if site1.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", site1.Language)
	}
	if site1.Domain != "" {
		t.Errorf("Expected domain '', got '%s'", site1.Domain)
	}
	if site1.RootPath != "/sitecore/content/asmblii/shared-config" {
		t.Errorf("Expected rootPath '/sitecore/content/asmblii/shared-config', got '%s'", site1.RootPath)
	}
	if site1.StartPath != "/Home" {
		t.Errorf("Expected startPath '/Home', got '%s'", site1.StartPath)
	}
	if site1.BrowserTitle != "Shared-config - Sitecore" {
		t.Errorf("Expected browserTitle 'Shared-config - Sitecore', got '%s'", site1.BrowserTitle)
	}
	if site1.CacheHtml != false {
		t.Errorf("Expected cacheHtml false, got %v", site1.CacheHtml)
	}
	if site1.CacheMedia != true {
		t.Errorf("Expected cacheMedia true, got %v", site1.CacheMedia)
	}
	if site1.EnablePreview != true {
		t.Errorf("Expected enablePreview true, got %v", site1.EnablePreview)
	}
	if site1.RootItem.ItemID != "59e05408e2c74916b27ca718bbabf4d9" {
		t.Errorf("Expected rootItem.itemId '59e05408e2c74916b27ca718bbabf4d9', got '%s'", site1.RootItem.ItemID)
	}
	if site1.StartItem.ItemID != "59e05408e2c74916b27ca718bbabf4d9" {
		t.Errorf("Expected startItem.itemId '59e05408e2c74916b27ca718bbabf4d9', got '%s'", site1.StartItem.ItemID)
	}
	if site1.StartItem.Path != "/sitecore/content/asmblii/shared-config" {
		t.Errorf("Expected startItem.path '/sitecore/content/asmblii/shared-config', got '%s'", site1.StartItem.Path)
	}

	// Test second site (reference) - specifically test the complex hostName with pipes
	site2 := sitesResponse.Sites[1]
	if site2.Name != "reference" {
		t.Errorf("Expected site name 'reference', got '%s'", site2.Name)
	}
	if site2.HostName != "dev-website-reference.vercel.app|dev-website-reference*.vercel.app|localhost:3000" {
		t.Errorf("Expected hostName 'dev-website-reference.vercel.app|dev-website-reference*.vercel.app|localhost:3000', got '%s'", site2.HostName)
	}

	// Test third site (website)
	site3 := sitesResponse.Sites[2]
	if site3.Name != "website" {
		t.Errorf("Expected site name 'website', got '%s'", site3.Name)
	}
	if site3.HostName != "" {
		t.Errorf("Expected hostName '', got '%s'", site3.HostName)
	}
	if site3.Domain != "extranet" {
		t.Errorf("Expected domain 'extranet', got '%s'", site3.Domain)
	}
}

func TestParseSiteByNameResponse(t *testing.T) {
	// Sample response data for single site
	responseData := map[string]interface{}{
		"site": map[string]interface{}{
			"name":            "shared-config",
			"hostName":        "*",
			"targetHostName":  "",
			"contentLanguage": map[string]interface{}{"name": "en"},
			"language":        "en",
			"domain":          nil,
			"rootPath":        "/sitecore/content/asmblii/shared-config",
			"startPath":       "/Home",
			"browserTitle":    "Shared-config - Sitecore",
			"cacheHtml":       false,
			"cacheMedia":      true,
			"enablePreview":   true,
			"rootItem":        map[string]interface{}{"itemId": "59e05408e2c74916b27ca718bbabf4d9"},
			"startItem":       map[string]interface{}{"itemId": "59e05408e2c74916b27ca718bbabf4d9", "path": "/sitecore/content/asmblii/shared-config"},
		},
	}

	var result struct {
		Site *Site `json:"site"`
	}
	err := parseGraphQLResponse(responseData, &result)
	if err != nil {
		t.Fatalf("Failed to parse site response: %v", err)
	}

	if result.Site == nil {
		t.Fatal("Expected site to be non-nil")
	}

	site := result.Site
	if site.Name != "shared-config" {
		t.Errorf("Expected site name 'shared-config', got '%s'", site.Name)
	}
	if site.HostName != "*" {
		t.Errorf("Expected hostName '*', got '%s'", site.HostName)
	}
	if site.TargetHostName != "" {
		t.Errorf("Expected targetHostName '', got '%s'", site.TargetHostName)
	}
}
