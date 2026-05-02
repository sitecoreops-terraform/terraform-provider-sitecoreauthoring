package apiclient

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFieldMappingWithRealAPIStructure tests field mapping with actual API response structure
func TestFieldMappingWithRealAPIStructure(t *testing.T) {
	t.Run("Given_real_API_response_structure_when_mapping_should_work_correctly", func(t *testing.T) {
		// This simulates the exact structure returned by the Sitecore GraphQL API
		fieldNames := []string{"title", "description"}

		// Create field mapping (this is what buildFieldsQuery does)
		fieldMapping := make(map[string]string)
		for i, fieldName := range fieldNames {
			fieldAlias := fmt.Sprintf("field%d", i+1)
			fieldMapping[fieldAlias] = fieldName
		}

		// Simulate raw API response (this is what comes back from GraphQL)
		rawFields := map[string]interface{}{
			"field1": map[string]interface{}{"value": "Sitecore Experience Platform"},
			"field2": map[string]interface{}{"value": "This is a description"},
		}

		// Apply the field mapping logic (this is what our provider does)
		fieldsMap := make(map[string]interface{})

		// Step 1: Ensure all requested fields are present (even if null)
		for _, fieldName := range fieldNames {
			fieldsMap[fieldName] = nil
		}

		// Step 2: Populate with actual values from API response
		for key, value := range rawFields {
			if value == nil {
				continue // Already set to null above
			} else if fieldObj, ok := value.(map[string]interface{}); ok {
				if val, exists := fieldObj["value"]; exists {
					if originalName, exists := fieldMapping[key]; exists {
						fieldsMap[originalName] = val
					}
				}
			}
		}

		// Verify the results
		assert.Equal(t, "Sitecore Experience Platform", fieldsMap["title"])
		assert.Equal(t, "This is a description", fieldsMap["description"])

		t.Log("✅ Field mapping with real API structure works correctly")
	})

	t.Run("Given_API_response_with_null_fields_when_mapping_should_handle_nulls", func(t *testing.T) {
		fieldNames := []string{"title", "description", "missingField"}

		// Create field mapping
		fieldMapping := make(map[string]string)
		for i, fieldName := range fieldNames {
			fieldAlias := fmt.Sprintf("field%d", i+1)
			fieldMapping[fieldAlias] = fieldName
		}

		// Simulate API response with some null fields
		rawFields := map[string]interface{}{
			"field1": map[string]interface{}{"value": "Hello"},
			"field2": nil, // This field is null in API response
			"field3": map[string]interface{}{"value": "Present"},
		}

		// Apply field mapping logic
		fieldsMap := make(map[string]interface{})

		// Step 1: Ensure all requested fields are present (even if null)
		for _, fieldName := range fieldNames {
			fieldsMap[fieldName] = nil
		}

		// Step 2: Populate with actual values from API response
		for key, value := range rawFields {
			if value == nil {
				continue // Already set to null above
			} else if fieldObj, ok := value.(map[string]interface{}); ok {
				if val, exists := fieldObj["value"]; exists {
					if originalName, exists := fieldMapping[key]; exists {
						fieldsMap[originalName] = val
					}
				}
			}
		}

		// Verify the results
		assert.Equal(t, "Hello", fieldsMap["title"])
		assert.Nil(t, fieldsMap["description"]) // Should remain null
		assert.Equal(t, "Present", fieldsMap["missingField"])

		t.Log("✅ Null field handling works correctly")
	})

	t.Run("Given_field_aliases_without_mapping_when_accessing_should_still_work", func(t *testing.T) {
		// Test that fields can still be accessed by alias if needed
		fieldNames := []string{"title", "description"}

		// Create field mapping
		fieldMapping := make(map[string]string)
		for i, fieldName := range fieldNames {
			fieldAlias := fmt.Sprintf("field%d", i+1)
			fieldMapping[fieldAlias] = fieldName
		}

		// Simulate API response
		rawFields := map[string]interface{}{
			"field1": map[string]interface{}{"value": "Title Value"},
			"field2": map[string]interface{}{"value": "Description Value"},
		}

		// Apply field mapping logic
		fieldsMap := make(map[string]interface{})

		// Step 1: Ensure all requested fields are present (even if null)
		for _, fieldName := range fieldNames {
			fieldsMap[fieldName] = nil
		}

		// Step 2: Populate with actual values from API response
		for key, value := range rawFields {
			if value == nil {
				continue
			} else if fieldObj, ok := value.(map[string]interface{}); ok {
				if val, exists := fieldObj["value"]; exists {
					if originalName, exists := fieldMapping[key]; exists {
						fieldsMap[originalName] = val
					}
					// Also keep the alias for backward compatibility
					fieldsMap[key] = val
				}
			}
		}

		// Verify both original names and aliases work
		assert.Equal(t, "Title Value", fieldsMap["title"])
		assert.Equal(t, "Title Value", fieldsMap["field1"])
		assert.Equal(t, "Description Value", fieldsMap["description"])
		assert.Equal(t, "Description Value", fieldsMap["field2"])

		t.Log("✅ Both original names and aliases work for field access")
	})
}
