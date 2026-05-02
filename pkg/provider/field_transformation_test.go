package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sitecoreops-terraform/terraform-provider-sitecoreauthoring/pkg/apiclient"
)

// TestFieldTransformationForChildren tests that field transformation works correctly for children
func TestFieldTransformationForChildren(t *testing.T) {
	t.Run("Given_GraphQL_field_structure_with_null_and_value_objects_when_transforming_should_return_uniform_string_values", func(t *testing.T) {
		// Test data that matches the actual API response
		rawFields := map[string]interface{}{
			"title": map[string]interface{}{
				"value": "Sitecore Experience Platform",
			},
			"description": nil,
			"tenantName": map[string]interface{}{
				"value": "ClydeCo",
			},
			"emptyField": map[string]interface{}{
				// Missing "value" property - should become nil
			},
		}

		// Apply transformation
		transformedFields := apiclient.TransformGraphQLFields(rawFields)

		// Verify transformation
		if len(transformedFields) != 4 {
			t.Errorf("Expected 4 fields after transformation, got %d", len(transformedFields))
		}

		// Check title field
		titleValue, ok := transformedFields["title"]
		if !ok {
			t.Error("Title field should exist after transformation")
		} else if titleValue != "Sitecore Experience Platform" {
			t.Errorf("Expected title 'Sitecore Experience Platform', got '%v'", titleValue)
		}

		// Check description field (should be nil)
		descriptionValue, ok := transformedFields["description"]
		if !ok {
			t.Error("Description field should exist after transformation")
		} else if descriptionValue != nil {
			t.Errorf("Expected description to be nil, got '%v'", descriptionValue)
		}

		// Check tenantName field
		tenantNameValue, ok := transformedFields["tenantName"]
		if !ok {
			t.Error("tenantName field should exist after transformation")
		} else if tenantNameValue != "ClydeCo" {
			t.Errorf("Expected tenantName 'ClydeCo', got '%v'", tenantNameValue)
		}

		// Check emptyField (missing value property)
		emptyFieldValue, ok := transformedFields["emptyField"]
		if !ok {
			t.Error("emptyField should exist after transformation")
		} else if emptyFieldValue != nil {
			t.Errorf("Expected emptyField to be nil, got '%v'", emptyFieldValue)
		}

		t.Log("✅ Field transformation works correctly for GraphQL structure")
	})

	t.Run("Given_transformed_fields_when_converting_to_Terraform_should_create_proper_map", func(t *testing.T) {
		// Transformed fields from previous test
		transformedFields := map[string]interface{}{
			"title":       "Sitecore Experience Platform",
			"description": nil,
			"tenantName":  "ClydeCo",
		}

		// Convert to Terraform map
		fieldsMap := make(map[string]types.String)
		for key, value := range transformedFields {
			if value == nil {
				fieldsMap[key] = types.StringNull()
			} else if strValue, ok := value.(string); ok {
				fieldsMap[key] = types.StringValue(strValue)
			} else {
				fieldsMap[key] = types.StringValue(fmt.Sprintf("%v", value))
			}
		}

		// Verify the map
		if len(fieldsMap) != 3 {
			t.Errorf("Expected 3 fields in Terraform map, got %d", len(fieldsMap))
		}

		// Check title field
		titleAttr, ok := fieldsMap["title"]
		if !ok {
			t.Error("Title field should exist in Terraform map")
		} else if !titleAttr.Equal(types.StringValue("Sitecore Experience Platform")) {
			t.Errorf("Expected title StringValue, got '%v'", titleAttr)
		}

		// Check description field (should be null)
		descriptionAttr, ok := fieldsMap["description"]
		if !ok {
			t.Error("Description field should exist in Terraform map")
		} else if !descriptionAttr.IsNull() {
			t.Errorf("Expected description to be null, got '%v'", descriptionAttr)
		}

		// Check tenantName field
		tenantNameAttr, ok := fieldsMap["tenantName"]
		if !ok {
			t.Error("tenantName field should exist in Terraform map")
		} else if !tenantNameAttr.Equal(types.StringValue("ClydeCo")) {
			t.Errorf("Expected tenantName StringValue, got '%v'", tenantNameAttr)
		}

		t.Log("✅ Terraform field map creation works correctly")
	})
}
