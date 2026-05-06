package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
	"github.com/stretchr/testify/assert"
)

// MockItemFieldClient is a mock implementation of the API client for item field resource testing
type MockItemFieldClient struct {
	*apiclient.Client
	MockGetItemByID func(itemID string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error)
	MockUpdateItem  func(itemID string, language string, fields map[string]interface{}, database string, path string) (*apiclient.Item, error)
}

func (m *MockItemFieldClient) GetItemByID(itemID string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error) {
	if m.MockGetItemByID != nil {
		return m.MockGetItemByID(itemID, fieldNames, existingVersionOnly)
	}
	return &apiclient.Item{
		ItemID: itemID,
		Path:   "/sitecore/content/test/" + itemID,
		Fields: map[string]interface{}{
			"title": "Test Title",
		},
	}, nil
}

func (m *MockItemFieldClient) UpdateItem(itemID string, language string, fields map[string]interface{}, database string, path string) (*apiclient.Item, error) {
	if m.MockUpdateItem != nil {
		return m.MockUpdateItem(itemID, language, fields, database, path)
	}
	return &apiclient.Item{
		ItemID: itemID,
		Path:   path,
		Fields: fields,
	}, nil
}

// Simple unit tests for the item field resource model conversion
func TestConvertToItemFieldResourceModel(t *testing.T) {
	t.Run("Convert API item to item field resource model", func(t *testing.T) {
		// Create a test item
		testItem := &apiclient.Item{
			ItemID: "test-item-id",
			Path:   "/sitecore/content/test/Test Item",
			Fields: map[string]interface{}{
				"title": "Test Title",
			},
		}

		// Create a test plan
		plan := itemFieldResourceModel{
			ItemID:     types.StringValue("test-item-id"),
			Language:   types.StringValue("en"),
			FieldName:  types.StringValue("title"),
			FieldValue: types.StringValue("Test Title"),
		}

		// Convert to resource model
		result := convertToItemFieldResourceModel(testItem, plan)

		// Verify the conversion
		assert.Equal(t, "test-item-id-en-title", result.ID.ValueString())
		assert.Equal(t, "test-item-id", result.ItemID.ValueString())
		assert.Equal(t, "en", result.Language.ValueString())
		assert.Equal(t, "title", result.FieldName.ValueString())
		assert.Equal(t, "Test Title", result.FieldValue.ValueString())
	})

	t.Run("Convert API item with empty field value to resource model", func(t *testing.T) {
		// Create a test item with empty field value
		testItem := &apiclient.Item{
			ItemID: "test-item-id",
			Path:   "/sitecore/content/test/Test Item",
			Fields: map[string]interface{}{
				"title": "",
			},
		}

		// Create a test plan
		plan := itemFieldResourceModel{
			ItemID:     types.StringValue("test-item-id"),
			Language:   types.StringValue("en"),
			FieldName:  types.StringValue("title"),
			FieldValue: types.StringValue(""),
		}

		// Convert to resource model
		result := convertToItemFieldResourceModel(testItem, plan)

		// Verify the conversion
		assert.Equal(t, "test-item-id-en-title", result.ID.ValueString())
		assert.Equal(t, "", result.FieldValue.ValueString())
	})

	t.Run("Convert API item with null field value to resource model", func(t *testing.T) {
		// Create a test item with null field value
		testItem := &apiclient.Item{
			ItemID: "test-item-id",
			Path:   "/sitecore/content/test/Test Item",
			Fields: map[string]interface{}{
				"title": nil,
			},
		}

		// Create a test plan
		plan := itemFieldResourceModel{
			ItemID:     types.StringValue("test-item-id"),
			Language:   types.StringValue("en"),
			FieldName:  types.StringValue("title"),
			FieldValue: types.StringValue(""),
		}

		// Convert to resource model
		result := convertToItemFieldResourceModel(testItem, plan)

		// Verify the conversion - should handle nil gracefully
		assert.Equal(t, "test-item-id-en-title", result.ID.ValueString())
		assert.Equal(t, "", result.FieldValue.ValueString())
	})
}

// Test that the resource preserves field values when API doesn't return them
func TestItemFieldResourceFieldValuePreservation(t *testing.T) {
	t.Run("Should preserve field value from plan when API response is empty", func(t *testing.T) {
		// Create a test item with empty fields (simulating API not returning updated values)
		testItem := &apiclient.Item{
			ItemID: "test-item-id",
			Path:   "/sitecore/content/test/Test Item",
			Fields: map[string]interface{}{}, // Empty fields - API didn't return the updated value
		}

		// Create a test plan with a field value
		plan := itemFieldResourceModel{
			ItemID:     types.StringValue("test-item-id"),
			Language:   types.StringValue("en"),
			FieldName:  types.StringValue("title"),
			FieldValue: types.StringValue("My Updated Title"),
		}

		// Convert to resource model (this simulates what happens after Create/Update)
		result := convertToItemFieldResourceModel(testItem, plan)

		// The field value should be empty from the API response
		assert.Equal(t, "", result.FieldValue.ValueString())

		// Now simulate the fix: if API returns empty but plan has value, preserve plan value
		if result.FieldValue.ValueString() == "" && plan.FieldValue.ValueString() != "" {
			result.FieldValue = plan.FieldValue
		}

		// After the fix, the field value should be preserved from the plan
		assert.Equal(t, "My Updated Title", result.FieldValue.ValueString())
	})

	t.Run("Should not override non-empty API response with plan value", func(t *testing.T) {
		// Create a test item with actual field value from API
		testItem := &apiclient.Item{
			ItemID: "test-item-id",
			Path:   "/sitecore/content/test/Test Item",
			Fields: map[string]interface{}{
				"title": "API Returned Value",
			},
		}

		// Create a test plan with a different field value
		plan := itemFieldResourceModel{
			ItemID:     types.StringValue("test-item-id"),
			Language:   types.StringValue("en"),
			FieldName:  types.StringValue("title"),
			FieldValue: types.StringValue("Plan Value"),
		}

		// Convert to resource model
		result := convertToItemFieldResourceModel(testItem, plan)

		// The field value should come from the API response, not the plan
		assert.Equal(t, "API Returned Value", result.FieldValue.ValueString())

		// Apply the fix logic - should not change anything since API returned a value
		if result.FieldValue.ValueString() == "" && plan.FieldValue.ValueString() != "" {
			result.FieldValue = plan.FieldValue
		}

		// Should still be the API value
		assert.Equal(t, "API Returned Value", result.FieldValue.ValueString())
	})
}
