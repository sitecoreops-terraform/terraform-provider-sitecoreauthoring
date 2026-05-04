package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
	"github.com/stretchr/testify/assert"
)

// MockResourceClient is a mock implementation of the API client for resource testing
type MockResourceClient struct {
	*apiclient.Client
	MockCreateItem              func(name, templateID, parentID, language string, fields map[string]interface{}) (*apiclient.Item, error)
	MockGetItemByPath           func(path string, existingVersionOnly *bool) (*apiclient.Item, error)
	MockGetItemByID             func(itemID string, existingVersionOnly *bool) (*apiclient.Item, error)
	MockUpdateItem              func(itemID, language string, version int, fields map[string]interface{}, database, path string) (*apiclient.Item, error)
	MockDeleteItem              func(path string, permanently bool) (bool, error)
	MockGetItemByPathWithFields func(path string, existingVersionOnly *bool) (*apiclient.Item, error)
	MockGetItemByIDWithFields   func(itemID string, existingVersionOnly *bool) (*apiclient.Item, error)
}

func (m *MockResourceClient) CreateItem(name, templateID, parentID, language string, fields map[string]interface{}) (*apiclient.Item, error) {
	if m.MockCreateItem != nil {
		return m.MockCreateItem(name, templateID, parentID, language, fields)
	}
	return &apiclient.Item{
		ItemID:     "test-item-id",
		Name:       name,
		Path:       "/sitecore/content/test/" + name,
		TemplateID: templateID,
		Fields:     fields,
	}, nil
}

func (m *MockResourceClient) GetItemByPath(path string, existingVersionOnly *bool) (*apiclient.Item, error) {
	if m.MockGetItemByPath != nil {
		return m.MockGetItemByPath(path, existingVersionOnly)
	}
	return &apiclient.Item{
		ItemID:     "test-item-id",
		Name:       "Test Item",
		Path:       path,
		TemplateID: "{TEMPLATE-ID}",
		Fields: map[string]interface{}{
			"title": "Test Title",
			"text":  "Test Text",
		},
	}, nil
}

func (m *MockResourceClient) GetItemByID(itemID string, existingVersionOnly *bool) (*apiclient.Item, error) {
	if m.MockGetItemByID != nil {
		return m.MockGetItemByID(itemID, existingVersionOnly)
	}
	return &apiclient.Item{
		ItemID:     itemID,
		Name:       "Test Item",
		Path:       "/sitecore/content/test/" + itemID,
		TemplateID: "{TEMPLATE-ID}",
		Fields: map[string]interface{}{
			"title": "Test Title",
			"text":  "Test Text",
		},
	}, nil
}

func (m *MockResourceClient) UpdateItem(itemID, language string, version int, fields map[string]interface{}, database, path string) (*apiclient.Item, error) {
	if m.MockUpdateItem != nil {
		return m.MockUpdateItem(itemID, language, version, fields, database, path)
	}
	return &apiclient.Item{
		ItemID:     itemID,
		Name:       "Updated Item",
		Path:       path,
		TemplateID: "{TEMPLATE-ID}",
		Fields:     fields,
	}, nil
}

func (m *MockResourceClient) DeleteItem(path string, permanently bool) (bool, error) {
	if m.MockDeleteItem != nil {
		return m.MockDeleteItem(path, permanently)
	}
	return true, nil
}

func (m *MockResourceClient) GetItemByPathWithFields(path string, existingVersionOnly *bool) (*apiclient.Item, error) {
	if m.MockGetItemByPathWithFields != nil {
		return m.MockGetItemByPathWithFields(path, existingVersionOnly)
	}
	return &apiclient.Item{
		ItemID:     "test-item-id",
		Name:       "Test Item",
		Path:       path,
		TemplateID: "{TEMPLATE-ID}",
		Fields: map[string]interface{}{
			"title": "Test Title",
			"text":  "Test Text",
		},
	}, nil
}

func (m *MockResourceClient) GetItemByIDWithFields(itemID string, existingVersionOnly *bool) (*apiclient.Item, error) {
	if m.MockGetItemByIDWithFields != nil {
		return m.MockGetItemByIDWithFields(itemID, existingVersionOnly)
	}
	return &apiclient.Item{
		ItemID:     itemID,
		Name:       "Test Item",
		Path:       "/sitecore/content/test/" + itemID,
		TemplateID: "{TEMPLATE-ID}",
		Fields: map[string]interface{}{
			"title": "Test Title",
			"text":  "Test Text",
		},
	}, nil
}

// Simple unit tests for the resource model conversion
func TestConvertToResourceModel(t *testing.T) {
	t.Run("Convert API item to resource model", func(t *testing.T) {
		// Create a test item
		testItem := &apiclient.Item{
			ItemID:     "test-item-id",
			Name:       "Test Item",
			Path:       "/sitecore/content/test/Test Item",
			TemplateID: "{TEMPLATE-ID}",
			Fields: map[string]interface{}{
				"title": "Test Title",
				"text":  "Test Text",
			},
		}

		// Create a test plan
		plan := itemResourceModel{
			Name:       types.StringValue("Test Item"),
			ParentID:   types.StringValue("{PARENT-ID}"),
			TemplateID: types.StringValue("{TEMPLATE-ID}"),
			Language:   types.StringValue("en"),
		}

		// Convert to resource model
		result := convertToResourceModel(testItem, plan)

		// Verify the conversion
		if result.ItemID.ValueString() != "test-item-id" {
			t.Errorf("Expected item_id 'test-item-id', got '%s'", result.ItemID.ValueString())
		}
		if result.Name.ValueString() != "Test Item" {
			t.Errorf("Expected name 'Test Item', got '%s'", result.Name.ValueString())
		}
		if result.Path.ValueString() != "/sitecore/content/test/Test Item" {
			t.Errorf("Expected path '/sitecore/content/test/Test Item', got '%s'", result.Path.ValueString())
		}

		// Check fields
		if len(result.Fields.Elements()) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(result.Fields.Elements()))
		}
	})

	t.Run("Convert API item with null fields to resource model", func(t *testing.T) {
		// Create a test item with no fields
		testItem := &apiclient.Item{
			ItemID:     "test-item-id",
			Name:       "Test Item",
			Path:       "/sitecore/content/test/Test Item",
			TemplateID: "{TEMPLATE-ID}",
			Fields:     map[string]interface{}{},
		}

		// Create a test plan with null fields
		plan := itemResourceModel{
			Name:       types.StringValue("Test Item"),
			ParentID:   types.StringValue("{PARENT-ID}"),
			TemplateID: types.StringValue("{TEMPLATE-ID}"),
			Language:   types.StringValue("en"),
			Fields:     types.MapNull(types.StringType),
		}

		// Convert to resource model
		result := convertToResourceModel(testItem, plan)

		// Verify the conversion - fields should remain null to maintain consistency
		if !result.Fields.IsNull() {
			t.Errorf("Expected fields to be null, but got %v", result.Fields)
		}
	})

	t.Run("Convert API item with empty fields to resource model", func(t *testing.T) {
		// Create a test item with empty fields
		testItem := &apiclient.Item{
			ItemID:     "test-item-id",
			Name:       "Test Item",
			Path:       "/sitecore/content/test/Test Item",
			TemplateID: "{TEMPLATE-ID}",
			Fields:     map[string]interface{}{},
		}

		// Create a test plan with empty fields map
		plan := itemResourceModel{
			Name:       types.StringValue("Test Item"),
			ParentID:   types.StringValue("{PARENT-ID}"),
			TemplateID: types.StringValue("{TEMPLATE-ID}"),
			Language:   types.StringValue("en"),
			Fields:     types.MapValueMust(types.StringType, map[string]attr.Value{}),
		}

		// Convert to resource model
		result := convertToResourceModel(testItem, plan)

		// Verify the conversion - fields should be empty map
		assert.NotNil(t, result.Fields)
		assert.Len(t, result.Fields.Elements(), 0)
	})

	t.Run("Convert API item null fields should remain null after apply", func(t *testing.T) {
		// This test reproduces the exact scenario from the bug report:
		// "Error: Provider produced inconsistent result after apply
		//  When applying changes to sitecoreauthoring_item.setting, provider
		//  produced an unexpected new value: .fields: was null, but now
		//  cty.MapValEmpty(cty.String)."

		// Create a test item with no fields (simulating API response)
		testItem := &apiclient.Item{
			ItemID:     "test-item-id",
			Name:       "Test Item",
			Path:       "/sitecore/content/test/Test Item",
			TemplateID: "{TEMPLATE-ID}",
			Fields:     map[string]interface{}{}, // Empty fields from API
		}

		// Create a test plan with null fields (simulating user not specifying fields in Terraform)
		plan := itemResourceModel{
			Name:       types.StringValue("Test Item"),
			ParentID:   types.StringValue("{PARENT-ID}"),
			TemplateID: types.StringValue("{TEMPLATE-ID}"),
			Language:   types.StringValue("en"),
			Fields:     types.MapNull(types.StringType), // Null fields in plan
		}

		// Convert to resource model (simulating the apply phase)
		result := convertToResourceModel(testItem, plan)

		assert.True(t, result.Fields.IsNull())
	})
}
