package provider

import (
	"context"
	"testing"
)

func TestProviderDataSources(t *testing.T) {
	// Create provider instance
	p := &sitecoreProvider{version: "test"}

	// Get data sources
	dataSources := p.DataSources(context.Background())

	// Verify we have the expected data sources
	if len(dataSources) != 3 {
		t.Fatalf("Expected 3 data sources, got %d", len(dataSources))
	}

	// Create instances to verify they can be instantiated
	for _, dsFunc := range dataSources {
		ds := dsFunc()
		if ds == nil {
			t.Error("Data source function returned nil")
		}
	}

	// Verify the data sources are the expected types
	foundSites := false
	foundItem := false
	foundItems := false

	for _, dsFunc := range dataSources {
		ds := dsFunc()
		switch ds.(type) {
		case *sitesDataSource:
			foundSites = true
		case *itemDataSource:
			foundItem = true
		case *itemsDataSource:
			foundItems = true
		}
	}

	if !foundSites {
		t.Error("Sites data source not found")
	}

	if !foundItem {
		t.Error("Item data source not found")
	}

	if !foundItems {
		t.Error("Items data source not found")
	}
}
