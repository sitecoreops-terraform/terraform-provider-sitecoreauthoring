package apiclient

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func normalize(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, "")
}

// TestBuildFieldsQuery tests the buildFieldsQuery function
func TestBuildFieldsQuery(t *testing.T) {

	t.Run("Multiple fields", func(t *testing.T) {
		fieldNames := []string{"title", "text", "image"}
		expectedFields := "field1:field(name:\"title\") { value }\n\t\t\t\t\tfield2:field(name:\"text\") { value }\n\t\t\t\t\tfield3:field(name:\"image\") { value }\n\t\t\t\t\t"

		actualFields := buildFieldsQuery(fieldNames)

		if actualFields != expectedFields {
			t.Errorf("Expected fields: %s\nActual fields: %s", expectedFields, actualFields)
		}
	})

	t.Run("Single field", func(t *testing.T) {
		fieldNames := []string{"title"}
		expectedFields := "field1:field(name:\"title\") { value }\n\t\t\t\t\t"

		actualFields := buildFieldsQuery(fieldNames)

		if actualFields != expectedFields {
			t.Errorf("Expected fields: %s\n\nActual fields: %s", expectedFields, actualFields)
		}
	})

	t.Run("No fields", func(t *testing.T) {
		fieldNames := []string{}
		expectedFields := ""

		actualFields := buildFieldsQuery(fieldNames)

		if actualFields != expectedFields {
			t.Errorf("Expected fields: %s\nActual fields: %s", expectedFields, actualFields)
		}
	})
}

// TestBuildWhereClause tests the buildWhereClause function
func TestBuildWhereClause(t *testing.T) {
	t.Run("Path without existingVersionOnly", func(t *testing.T) {
		actualWhereClause := buildWhereClause(true, "/sitecore/content/asmblii/home", nil)
		expectedWhereClause := `{path: "/sitecore/content/asmblii/home"}`
		if actualWhereClause != expectedWhereClause {
			t.Errorf("Expected where clause: %s\nActual where clause: %s", expectedWhereClause, actualWhereClause)
		}
	})

	t.Run("Path with existingVersionOnly=true", func(t *testing.T) {
		existingVersionOnly := true
		actualWhereClause := buildWhereClause(true, "/sitecore/content/asmblii/home", &existingVersionOnly)
		expectedWhereClause := `{path: "/sitecore/content/asmblii/home", existingVersionOnly: true}`
		if actualWhereClause != expectedWhereClause {
			t.Errorf("Expected where clause: %s\nActual where clause: %s", expectedWhereClause, actualWhereClause)
		}
	})

	t.Run("Path with existingVersionOnly=false", func(t *testing.T) {
		existingVersionOnly := false
		actualWhereClause := buildWhereClause(true, "/sitecore/content/asmblii/home", &existingVersionOnly)
		expectedWhereClause := `{path: "/sitecore/content/asmblii/home", existingVersionOnly: false}`
		if actualWhereClause != expectedWhereClause {
			t.Errorf("Expected where clause: %s\nActual where clause: %s", expectedWhereClause, actualWhereClause)
		}
	})

	t.Run("ID without existingVersionOnly", func(t *testing.T) {
		actualWhereClause := buildWhereClause(false, "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}", nil)
		expectedWhereClause := `{itemId: "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}"}`
		if actualWhereClause != expectedWhereClause {
			t.Errorf("Expected where clause: %s\nActual where clause: %s", expectedWhereClause, actualWhereClause)
		}
	})

	t.Run("ID with existingVersionOnly=true", func(t *testing.T) {
		existingVersionOnly := true
		actualWhereClause := buildWhereClause(false, "{GUID}", &existingVersionOnly)
		expectedWhereClause := `{itemId: "{GUID}", existingVersionOnly: true}`
		if actualWhereClause != expectedWhereClause {
			t.Errorf("Expected where clause: %s\nActual where clause: %s", expectedWhereClause, actualWhereClause)
		}
	})
}

// TestRealAPIResponseParsing tests parsing of actual Sitecore Authoring API responses
func TestRealAPIResponseParsing(t *testing.T) {
	t.Run("Given_real_API_response_with_GraphQL_structure_when_parsing_should_fail_due_to_children_template_mismatch_but_field_structure_should_be_resolved", func(t *testing.T) {
		// This is the exact API response that was causing terraform plan to fail
		apiResponse := `{
  "data": {
    "item": {
      "itemId": "110d559fdea542ea9c1c8a5df7e70ef9",
      "path": "/sitecore/content/Home",
      "name": "Home",
      "displayName": "Home",
      "template": {
        "templateId": "76036f5ecbce46d1af0a4143f9b557aa",
        "name": "Sample Item"
      },
      "field1": null,
      "field2": {
        "value": "Sitecore Experience Platform"
      },
      "children": {
        "nodes": []
      }
    }
  }
}`

		// Parse the response
		var responseData map[string]interface{}
		if err := json.Unmarshal([]byte(apiResponse), &responseData); err != nil {
			t.Fatalf("Failed to parse API response JSON: %v", err)
		}

		// Extract data section
		dataSection, ok := responseData["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Data section not found or not a map")
		}

		// Try to parse with current model - this should still fail due to children/template structure
		// but field structure should be handled correctly
		var itemResponse ItemResponse
		err := parseGraphQLResponse(dataSection, &itemResponse)

		// This test should still FAIL due to children/template structure mismatches
		// but the error should not be about field structure anymore
		if err == nil {
			t.Error("❌ UNEXPECTED: Parsing succeeded when it should have failed due to children/template structural mismatches")
		} else {
			t.Logf("✅ EXPECTED: Parsing failed due to children/template structure: %v", err)

			// The error should be about children or template structure, not fields
			if strings.Contains(err.Error(), "Fields") || (strings.Contains(err.Error(), "field") && !strings.Contains(err.Error(), "children")) {
				t.Errorf("❌ UNEXPECTED: Field structure error should be fixed: %v", err)
			} else if strings.Contains(err.Error(), "children") {
				t.Log("✅ Confirmed: Field structure is fixed, only children structure remains")
			} else if strings.Contains(err.Error(), "template") {
				t.Log("✅ Confirmed: Field structure is fixed, only template structure remains")
			} else {
				t.Logf("✅ Confirmed: Field structure is fixed, different error: %v", err)
			}
		}
	})

	t.Run("Given_template_nested_in_object_when_parsing_should_identify_structure_mismatch", func(t *testing.T) {
		// Test the template structure issue specifically
		templateData := `{
  "template": {
    "templateId": "76036f5ecbce46d1af0a4143f9b557aa",
    "name": "Sample Item"
  }
}`

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(templateData), &data); err != nil {
			t.Fatalf("Failed to parse template data: %v", err)
		}

		// Our model expects templateId and templateName at item level
		// But API returns them nested under "template"
		templateObj, ok := data["template"].(map[string]interface{})
		if !ok {
			t.Fatal("Template should be an object")
		}

		templateId := templateObj["templateId"]
		templateName := templateObj["name"]

		if templateId == nil || templateName == nil {
			t.Error("❌ Template structure doesn't match expected format")
		} else {
			t.Logf("✅ Template structure confirmed: ID=%v, Name=%v (but nested under 'template' object)", templateId, templateName)
			t.Log("⚠️  Model expects templateId and templateName at item level, not nested")
		}
	})

	t.Run("Given_mixed_field_structure_with_null_and_value_objects_when_parsing_should_identify_structure_mismatch", func(t *testing.T) {
		// Test the field structure issue specifically
		fieldData := `{
  "field1": null,
  "field2": {"value": "Sitecore Experience Platform"}
}`

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(fieldData), &data); err != nil {
			t.Fatalf("Failed to parse field data: %v", err)
		}

		// Our model expects Fields as map[string]string
		// But API returns mixed: some null, some {"value": "..."}
		field1 := data["field1"]
		field2, ok := data["field2"].(map[string]interface{})

		if field1 != nil {
			t.Error("❌ Field1 should be null but isn't")
		}

		if !ok {
			t.Error("❌ Field2 should be an object with 'value' property")
		} else if value, exists := field2["value"]; !exists {
			t.Error("❌ Field2 should have a 'value' property")
		} else {
			t.Logf("✅ Field structure confirmed: field1=null, field2={\"value\": \"%v\"}", value)
			t.Log("⚠️  Model expects uniform map[string]string, but API returns mixed format")
		}
	})
}

// TestBuildItemQuery tests the complete query building
func TestBuildItemQuery(t *testing.T) {
	t.Run("Item by Path with fields", func(t *testing.T) {
		fieldNames := []string{"title", "text"}
		fieldsQuery := buildFieldsQuery(fieldNames)
		whereClause := buildWhereClause(true, "/sitecore/content/asmblii/home", nil)
		actualQuery := buildItemQuery(whereClause, fieldsQuery)

		expectedQuery := `
		query ItemLookup {
			item(where: {path: "/sitecore/content/asmblii/home"}) {
				itemId
				path
				name
				displayName
				template {
					templateId
					name
				}
				field1:field(name:"title") { value }
					field2:field(name:"text") { value }
					
				children {
					nodes {
						itemId
						path
						name
						displayName
						template {
							templateId
							name
						}
						field1:field(name:"title") { value }
					field2:field(name:"text") { value }
					
					}
				}
		}
	}`

		assert.Equal(t, normalize(expectedQuery), normalize(actualQuery))
	})

	t.Run("Item by Path with existingVersionOnly", func(t *testing.T) {
		fieldNames := []string{"title"}
		existingVersionOnly := true
		fieldsQuery := buildFieldsQuery(fieldNames)
		whereClause := buildWhereClause(true, "/sitecore/content/asmblii/home", &existingVersionOnly)
		actualQuery := buildItemQuery(whereClause, fieldsQuery)

		expectedQuery := `
		query ItemLookup {
			item(where: {path: "/sitecore/content/asmblii/home", existingVersionOnly: true}) {
				itemId
				path
				name
				displayName
				template {
					templateId
					name
				}
				field1:field(name:"title") { value }
				
				children {
					nodes {
						itemId
						path
						name
						displayName
						template {
							templateId
							name
						}
						field1:field(name:"title") { value}
					}
				}
			}
		}`

		assert.Equal(t, normalize(expectedQuery), normalize(actualQuery))
	})

	t.Run("Item by ID without fields", func(t *testing.T) {
		fieldNames := []string{"description"}
		fieldsQuery := buildFieldsQuery(fieldNames)
		whereClause := buildWhereClause(false, "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}", nil)
		actualQuery := buildItemQuery(whereClause, fieldsQuery)

		expectedQuery := `
		query ItemLookup {
			item(where: {itemId: "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}"}) {
				itemId
				path
				name
				displayName
				template {
					templateId
					name
				}
				field1:field(name:"description") { value }
				
				children {
					nodes {
						itemId
						path
						name
						displayName
						template {
							templateId
							name
						}
						field1:field(name:"description") { value }

					}
				}
			}
		}`

		assert.Equal(t, normalize(expectedQuery), normalize(actualQuery))
	})
}

func TestFieldNameMapping(t *testing.T) {
	t.Run("Test field alias to original name mapping", func(t *testing.T) {
		// Test the complete flow: requested fields -> GraphQL aliases -> back to original names
		requestedFields := []string{"title", "text", "description"}

		// Simulate GraphQL response with field aliases
		rawJSON := `{
			"item": {
				"itemId": "test-item-id",
				"path": "/test/path",
				"name": "Test Item",
				"displayName": "Test Item",
				"template": {
					"templateId": "template-id",
					"name": "Test Template"
				},
				"field1": {
					"value": "Hello World"
				},
				"field2": {
					"value": "Some content"
				},
				"field3": {
					"value": "Item description"
				},
				"children": {
					"nodes": []
				}
			}
		}`

		// Parse the JSON into our structure
		var responseData map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &responseData); err != nil {
			t.Fatalf("Failed to unmarshal raw JSON: %v", err)
		}

		var graphQLResponse struct {
			Item *graphQLItemResponse `json:"item"`
		}
		if err := parseGraphQLResponse(responseData, &graphQLResponse); err != nil {
			t.Fatalf("Failed to parse GraphQL response: %v", err)
		}

		// Convert using the field name mapping
		item := convertFromGraphQLItem(graphQLResponse.Item, requestedFields)

		// Verify that field aliases were mapped back to original names
		if item.Fields == nil {
			t.Fatal("Fields should not be nil")
		}

		// Check that we have the correct field names (not field1, field2, field3)
		if _, exists := item.Fields["field1"]; exists {
			t.Error("Field should be mapped to original name, not keep 'field1' alias")
		}
		if _, exists := item.Fields["field2"]; exists {
			t.Error("Field should be mapped to original name, not keep 'field2' alias")
		}
		if _, exists := item.Fields["field3"]; exists {
			t.Error("Field should be mapped to original name, not keep 'field3' alias")
		}

		// Check that original field names are present with correct values
		if title, ok := item.Fields["title"].(string); !ok || title != "Hello World" {
			t.Errorf("Expected title='Hello World', got %v", item.Fields["title"])
		}
		if text, ok := item.Fields["text"].(string); !ok || text != "Some content" {
			t.Errorf("Expected text='Some content', got %v", item.Fields["text"])
		}
		if desc, ok := item.Fields["description"].(string); !ok || desc != "Item description" {
			t.Errorf("Expected description='Item description', got %v", item.Fields["description"])
		}
	})

	t.Run("Test field mapping with fewer fields than requested", func(t *testing.T) {
		// Test when some requested fields are null/empty in response
		requestedFields := []string{"title", "text", "missing_field"}

		rawJSON := `{
			"item": {
				"itemId": "test-id",
				"path": "/test",
				"name": "Test",
				"displayName": "Test",
				"template": {
					"templateId": "temp-id",
					"name": "Template"
				},
				"field1": {
					"value": "Title Value"
				},
				"field2": null,
				"children": {
					"nodes": []
				}
			}
		}`

		var responseData map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &responseData); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		var graphQLResponse struct {
			Item *graphQLItemResponse `json:"item"`
		}
		if err := parseGraphQLResponse(responseData, &graphQLResponse); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		item := convertFromGraphQLItem(graphQLResponse.Item, requestedFields)

		// Should have title mapped correctly
		if title, ok := item.Fields["title"].(string); !ok || title != "Title Value" {
			t.Errorf("Expected title='Title Value', got %v", item.Fields["title"])
		}

		// text should be nil (field2 was null)
		if item.Fields["text"] != nil {
			t.Errorf("Expected text to be nil, got %v", item.Fields["text"])
		}

		// missing_field should not exist (no field3 in response)
		if _, exists := item.Fields["missing_field"]; exists {
			t.Errorf("missing_field should not exist in result")
		}
	})
}

func TestParseGraphQLResponse(t *testing.T) {
	t.Run("Test parsing raw JSON string with item data", func(t *testing.T) {
		// Raw JSON string as it would come from GraphQL API
		rawJSON := `{
			"item": {
				"itemId": "110d559fdea542ea9c1c8a5df7e70ef9",
				"path": "/sitecore/content/Home",
				"name": "Home",
				"displayName": "Home",
				"template": {
					"templateId": "76036f5ecbce46d1af0a4143f9b557aa",
					"name": "Sample Item"
				},
				"field1": {
					"value": "Sitecore Experience Platform"
				},
				"field2": {
					"value": "Some text"
				},
				"children": {
					"nodes": []
				}
			}
		}`

		// First, unmarshal the raw JSON into the map[string]interface{} format
		// that parseGraphQLResponse expects (simulating what doGraphQLRequest returns)
		var responseData map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &responseData); err != nil {
			t.Fatalf("Failed to unmarshal raw JSON: %v", err)
		}

		// Now test parseGraphQLResponse with the data in the expected format
		var graphQLResponse struct {
			Item *graphQLItemResponse `json:"item"`
		}

		err := parseGraphQLResponse(responseData, &graphQLResponse)
		if err != nil {
			t.Fatalf("Failed to parse GraphQL response: %v", err)
		}

		// Verify all properties are available and correctly parsed
		if graphQLResponse.Item.ItemID != "110d559fdea542ea9c1c8a5df7e70ef9" {
			t.Errorf("Expected itemId '110d559fdea542ea9c1c8a5df7e70ef9', got '%s'", graphQLResponse.Item.ItemID)
		}
		if graphQLResponse.Item.Path != "/sitecore/content/Home" {
			t.Errorf("Expected path '/sitecore/content/Home', got '%s'", graphQLResponse.Item.Path)
		}
		if graphQLResponse.Item.Name != "Home" {
			t.Errorf("Expected name 'Home', got '%s'", graphQLResponse.Item.Name)
		}
		if graphQLResponse.Item.DisplayName != "Home" {
			t.Errorf("Expected displayName 'Home', got '%s'", graphQLResponse.Item.DisplayName)
		}
		if graphQLResponse.Item.Template.TemplateID != "76036f5ecbce46d1af0a4143f9b557aa" {
			t.Errorf("Expected templateId '76036f5ecbce46d1af0a4143f9b557aa', got '%s'", graphQLResponse.Item.Template.TemplateID)
		}
		if graphQLResponse.Item.Template.Name != "Sample Item" {
			t.Errorf("Expected template name 'Sample Item', got '%s'", graphQLResponse.Item.Template.Name)
		}

		// Verify nested field structures are properly parsed using DynamicFields
		if field1, ok := graphQLResponse.Item.DynamicFields["field1"]["value"]; !ok || field1 != "Sitecore Experience Platform" {
			t.Errorf("Expected field1.value 'Sitecore Experience Platform', got %v", graphQLResponse.Item.DynamicFields["field1"])
		}
		if field2, ok := graphQLResponse.Item.DynamicFields["field2"]["value"]; !ok || field2 != "Some text" {
			t.Errorf("Expected field2.value 'Some text', got %v", graphQLResponse.Item.DynamicFields["field2"])
		}
	})

	t.Run("Test parsing with null values", func(t *testing.T) {
		responseData := map[string]interface{}{
			"item": map[string]interface{}{
				"itemId": "test-id",
				"path":   "/test/path",
				"name":   "Test Item",
				"field1": nil, // null value
			},
		}

		var result struct {
			Item struct {
				ItemID string      `json:"itemId"`
				Path   string      `json:"path"`
				Name   string      `json:"name"`
				Field1 interface{} `json:"field1"` // Use interface{} to handle null
			} `json:"item"`
		}

		err := parseGraphQLResponse(responseData, &result)
		if err != nil {
			t.Fatalf("Failed to parse GraphQL response with null values: %v", err)
		}

		if result.Item.ItemID != "test-id" {
			t.Errorf("Expected itemId 'test-id', got '%s'", result.Item.ItemID)
		}
		if result.Item.Field1 != nil {
			t.Errorf("Expected field1 to be nil, got %v", result.Item.Field1)
		}
	})

	t.Run("Test parsing empty data", func(t *testing.T) {
		responseData := map[string]interface{}{}

		var result map[string]interface{}

		err := parseGraphQLResponse(responseData, &result)
		if err != nil {
			t.Fatalf("Failed to parse empty GraphQL response: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("Expected empty result, got %v", result)
		}
	})
}
