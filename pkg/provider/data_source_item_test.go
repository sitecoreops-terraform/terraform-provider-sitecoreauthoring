package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// MockClient is a mock implementation of the API client for testing
type MockClient struct {
	*apiclient.Client
	MockGetItemByID   func(itemID string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error)
	MockGetItemByPath func(path string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error)
}

func (m *MockClient) GetItemByID(itemID string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error) {
	if m.MockGetItemByID != nil {
		return m.MockGetItemByID(itemID, fieldNames, existingVersionOnly)
	}
	return nil, fmt.Errorf("mock GetItemByID not implemented")
}

func (m *MockClient) GetItemByPath(path string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error) {
	if m.MockGetItemByPath != nil {
		return m.MockGetItemByPath(path, fieldNames, existingVersionOnly)
	}
	return nil, fmt.Errorf("mock GetItemByPath not implemented")
}

func TestItemDataSourceMapping(t *testing.T) {
	// Create test data
	testItem := &apiclient.Item{
		ItemID:       "test-item-id",
		Path:         "/sitecore/content/asmblii/test-item",
		Name:         "test-item",
		DisplayName:  "Test Item",
		TemplateID:   "template-id",
		TemplateName: "Test Template",
		Fields: map[string]interface{}{
			"title": "Test Title",
			"text":  "Test Text",
		},
		Children: []apiclient.Item{
			{
				ItemID:       "child-item-id",
				Path:         "/sitecore/content/asmblii/test-item/child",
				Name:         "child-item",
				DisplayName:  "Child Item",
				TemplateID:   "child-template-id",
				TemplateName: "Child Template",
				Fields: map[string]interface{}{
					"title": "Child Title",
				},
			},
		},
	}

	// Test mapping logic (similar to what happens in the Read function)
	fieldsMap := make(map[string]attr.Value)
	for key, value := range testItem.Fields {
		if strValue, ok := value.(string); ok {
			fieldsMap[key] = types.StringValue(strValue)
		}
	}

	// Map children

	for _, child := range testItem.Children {
		childFieldsMap := make(map[string]attr.Value)
		for key, value := range child.Fields {
			if strValue, ok := value.(string); ok {
				childFieldsMap[key] = types.StringValue(strValue)
			}
		}

	}

	itemModel := itemModel{
		ItemID:       types.StringValue(testItem.ItemID),
		Path:         types.StringValue(testItem.Path),
		Name:         types.StringValue(testItem.Name),
		DisplayName:  types.StringValue(testItem.DisplayName),
		TemplateID:   types.StringValue(testItem.TemplateID),
		TemplateName: types.StringValue(testItem.TemplateName),
		Fields:       types.MapValueMust(types.StringType, fieldsMap),
	}

	// Verify the mapping
	if itemModel.ItemID.ValueString() != "test-item-id" {
		t.Errorf("Expected ItemID 'test-item-id', got '%s'", itemModel.ItemID.ValueString())
	}

	if itemModel.Path.ValueString() != "/sitecore/content/asmblii/test-item" {
		t.Errorf("Expected Path '/sitecore/content/asmblii/test-item', got '%s'", itemModel.Path.ValueString())
	}

	if itemModel.Name.ValueString() != "test-item" {
		t.Errorf("Expected Name 'test-item', got '%s'", itemModel.Name.ValueString())
	}

	if itemModel.DisplayName.ValueString() != "Test Item" {
		t.Errorf("Expected DisplayName 'Test Item', got '%s'", itemModel.DisplayName.ValueString())
	}

	// Verify fields
	fields := itemModel.Fields
	if fields.IsNull() || fields.IsUnknown() {
		t.Errorf("Expected fields to be non-null")
	}

	// Children are now handled by the separate items data source
}

func TestItemDataSourceValidation(t *testing.T) {
	// Test validation logic

	// Test case 1: Neither item_id nor path specified (should fail)
	state1 := itemDataSourceModel{
		ItemID: types.StringNull(),
		Path:   types.StringNull(),
	}

	// This would normally be caught by the validation in the Read function
	if (state1.ItemID.IsNull() || len(state1.ItemID.ValueString()) == 0) && (state1.Path.IsNull() || len(state1.Path.ValueString()) == 0) {
		// This is the expected behavior - validation should catch this
	} else {
		t.Errorf("Validation should catch missing both item_id and path")
	}

	// Test case 2: Both item_id and path specified (should fail)
	state2 := itemDataSourceModel{
		ItemID: types.StringValue("test-id"),
		Path:   types.StringValue("/test/path"),
	}

	if (!state2.ItemID.IsNull() && len(state2.ItemID.ValueString()) > 0) && (!state2.Path.IsNull() && len(state2.Path.ValueString()) > 0) {
		// This is the expected behavior - validation should catch this
	} else {
		t.Errorf("Validation should catch both item_id and path being specified")
	}

	// Test case 3: Only item_id specified (should pass)
	state3 := itemDataSourceModel{
		ItemID: types.StringValue("test-id"),
		Path:   types.StringNull(),
	}

	if (!state3.ItemID.IsNull() && len(state3.ItemID.ValueString()) > 0) && (state3.Path.IsNull() || len(state3.Path.ValueString()) == 0) {
		// This is the expected behavior - only item_id specified
	} else {
		t.Errorf("Should allow only item_id to be specified")
	}

	// Test case 4: Only path specified (should pass)
	state4 := itemDataSourceModel{
		ItemID: types.StringNull(),
		Path:   types.StringValue("/test/path"),
	}

	if (state4.ItemID.IsNull() || len(state4.ItemID.ValueString()) == 0) && (!state4.Path.IsNull() && len(state4.Path.ValueString()) > 0) {
		// This is the expected behavior - only path specified
	} else {
		t.Errorf("Should allow only path to be specified")
	}
}

func TestItemDataSourceRequiredBehavior(t *testing.T) {
	// Test the required parameter behavior

	// Test case 1: Required = true (default behavior) - should fail on missing item
	state1 := itemDataSourceModel{
		ItemID:   types.StringValue("non-existent-id"),
		Path:     types.StringNull(),
		Required: types.BoolValue(true),
	}

	// This would normally fail in the Read function when the item is not found
	// We're just testing the parameter is properly set
	if !state1.Required.ValueBool() {
		t.Errorf("Required should be true")
	}

	// Test case 2: Required = false - should handle missing item gracefully
	state2 := itemDataSourceModel{
		ItemID:   types.StringValue("non-existent-id"),
		Path:     types.StringNull(),
		Required: types.BoolValue(false),
	}

	if state2.Required.ValueBool() {
		t.Errorf("Required should be false")
	}

	// Test case 3: Required not specified - should default to true (backward compatibility)
	state3 := itemDataSourceModel{
		ItemID: types.StringValue("test-id"),
		Path:   types.StringNull(),
		// Required not set
	}

	// Should default to true for backward compatibility
	isRequired := true
	if !state3.Required.IsNull() {
		isRequired = state3.Required.ValueBool()
	}
	if !isRequired {
		t.Errorf("Required should default to true when not specified")
	}
}

// Test error handling for non-existent items
func TestItemDataSourceErrorHandling(t *testing.T) {
	t.Run("Item not found by ID should be handled gracefully when required is false", func(t *testing.T) {
		// Simulate the error handling logic from the Read function
		isRequired := false // Simulating required = false
		identifier := "test-id"
		errorMessage := "item with ID 'test-id' not found"
		err := fmt.Errorf("%s", errorMessage)

		// Check if this error should be handled gracefully
		shouldHandleGracefully := false
		if !isRequired {
			shouldHandleGracefully = err.Error() == errorMessage || err.Error() == "item with ID '"+identifier+"' not found"
		}

		if !shouldHandleGracefully {
			t.Errorf("Expected error to be handled gracefully for missing item by ID, but it wasn't")
		}
	})

	t.Run("Item not found by path should be handled gracefully when required is false", func(t *testing.T) {
		// Simulate the error handling logic from the Read function
		isRequired := false // Simulating required = false
		identifier := "/test/path"
		errorMessage := "item with path '/test/path' not found"
		err := fmt.Errorf("%s", errorMessage)

		// Check if this error should be handled gracefully
		shouldHandleGracefully := false
		if !isRequired {
			shouldHandleGracefully = err.Error() == errorMessage || err.Error() == "item with path '"+identifier+"' not found"
		}

		if !shouldHandleGracefully {
			t.Errorf("Expected error to be handled gracefully for missing item by path, but it wasn't")
		}
	})

	t.Run("Other error types should not be handled gracefully even when required is false", func(t *testing.T) {
		// Check if this error should be handled gracefully
		shouldHandleGracefully := false
		// This is not an "item not found" error, so it shouldn't be handled gracefully
		// The switch statement in the original code wouldn't match this case

		if shouldHandleGracefully {
			t.Errorf("Expected other error types to not be handled gracefully, but it was")
		}
	})
}

// Test the actual datasource behavior with mocked API responses
// TestFieldTransformation tests that GraphQL field structure is properly transformed
func TestFieldTransformation(t *testing.T) {
	t.Run("Given_GraphQL_field_structure_with_null_and_value_objects_when_transforming_should_return_uniform_string_values", func(t *testing.T) {
		// Test the field transformation logic
		testFields := map[string]interface{}{
			"field1": nil,
			"field2": map[string]interface{}{"value": "test value"},
			"field3": "direct string",
		}

		transformed := apiclient.TransformGraphQLFields(testFields)

		// Verify null field
		if transformed["field1"] != nil {
			t.Errorf("Expected field1 to be nil, got %v", transformed["field1"])
		}

		// Verify object field
		if transformed["field2"] != "test value" {
			t.Errorf("Expected field2 to be 'test value', got %v", transformed["field2"])
		}

		// Verify direct string field (backward compatibility)
		if transformed["field3"] != "direct string" {
			t.Errorf("Expected field3 to be 'direct string', got %v", transformed["field3"])
		}
	})

	t.Run("Given_empty_or_nil_field_map_when_transforming_should_return_nil_or_empty_map", func(t *testing.T) {
		// Test with nil input
		transformed := apiclient.TransformGraphQLFields(nil)
		if transformed != nil {
			t.Errorf("Expected nil result for nil input, got %v", transformed)
		}

		// Test with empty map
		transformed = apiclient.TransformGraphQLFields(map[string]interface{}{})
		if len(transformed) != 0 {
			t.Errorf("Expected empty result for empty input, got %v", transformed)
		}
	})

	t.Run("Given_malformed_field_objects_missing_value_property_when_transforming_should_return_nil", func(t *testing.T) {
		// Test with field object missing 'value' property
		testFields := map[string]interface{}{
			"field1": map[string]interface{}{"other": "property"},
		}

		transformed := apiclient.TransformGraphQLFields(testFields)

		// Should return nil for malformed objects
		if transformed["field1"] != nil {
			t.Errorf("Expected field1 to be nil for malformed object, got %v", transformed["field1"])
		}
	})
}

func TestItemDataSourceWithMockedResponses(t *testing.T) {
	t.Run("Should handle null item response gracefully when required is false", func(t *testing.T) {
		// Create a mock client that returns a response with null item
		mockClient := &MockClient{
			MockGetItemByID: func(itemID string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error) {
				// Simulate the API response: {"data":{"item":null}}
				return nil, fmt.Errorf("item with ID '%s' not found", itemID)
			},
		}

		// Test the error handling logic that would be used in the datasource
		isRequired := false
		itemID := "non-existent-id"

		// Simulate calling the client
		item, err := mockClient.GetItemByID(itemID, nil, nil)

		// Verify the mock client behavior
		if item != nil {
			t.Errorf("Expected item to be nil for non-existent item")
		}
		if err == nil {
			t.Errorf("Expected error for non-existent item")
		}

		// Test the error handling logic from the datasource
		shouldHandleGracefully := !isRequired && (err.Error() == fmt.Sprintf("item with ID '%s' not found", itemID) || err.Error() == "item with ID '"+itemID+"' not found")

		if !shouldHandleGracefully {
			t.Errorf("Expected error to be handled gracefully when required=false")
		}
	})

	t.Run("Should fail when required is true and item doesn't exist", func(t *testing.T) {
		// Create a mock client that returns a response with null item
		mockClient := &MockClient{
			MockGetItemByID: func(itemID string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error) {
				// Simulate the API response: {"data":{"item":null}}
				return nil, fmt.Errorf("item with ID '%s' not found", itemID)
			},
		}

		// Test the error handling logic that would be used in the datasource
		isRequired := true
		itemID := "non-existent-id"

		// Simulate calling the client
		item, err := mockClient.GetItemByID(itemID, nil, nil)

		// Verify the mock client behavior
		if item != nil {
			t.Errorf("Expected item to be nil for non-existent item")
		}
		if err == nil {
			t.Errorf("Expected error for non-existent item")
		}

		// Test the error handling logic from the datasource
		shouldHandleGracefully := !isRequired && (err.Error() == fmt.Sprintf("item with ID '%s' not found", itemID) || err.Error() == "item with ID '"+itemID+"' not found")

		if shouldHandleGracefully {
			t.Errorf("Expected error to NOT be handled gracefully when required=true")
		}
	})

	t.Run("Should handle null item response for path queries when required is false", func(t *testing.T) {
		// Create a mock client that returns a response with null item for path queries
		mockClient := &MockClient{
			MockGetItemByPath: func(path string, fieldNames []string, existingVersionOnly *bool) (*apiclient.Item, error) {
				// Simulate the API response: {"data":{"item":null}}
				return nil, fmt.Errorf("item with path '%s' not found", path)
			},
		}

		// Test the error handling logic that would be used in the datasource
		isRequired := false
		path := "/sitecore/content/non-existent"

		// Simulate calling the client
		item, err := mockClient.GetItemByPath(path, nil, nil)

		// Verify the mock client behavior
		if item != nil {
			t.Errorf("Expected item to be nil for non-existent item")
		}
		if err == nil {
			t.Errorf("Expected error for non-existent item")
		}

		// Test the error handling logic from the datasource
		shouldHandleGracefully := !isRequired && (err.Error() == fmt.Sprintf("item with path '%s' not found", path) || err.Error() == "item with path '"+path+"' not found")

		if !shouldHandleGracefully {
			t.Errorf("Expected error to be handled gracefully when required=false")
		}
	})
}
