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

// TestGetItemQueryBuilderFields tests the GetItemQueryBuilder field building
func TestGetItemQueryBuilderFields(t *testing.T) {
	t.Run("Multiple fields", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.AddField("title", nil)
		builder.AddField("text", nil)
		builder.AddField("image", nil)

		// Extract just the fields part for testing
		fullQuery := builder.Build()
		// The fields should be in the query
		if !strings.Contains(fullQuery, `field1:field(name:"title") { value }`) {
			t.Errorf("Expected field1:field(name:\"title\") { value } in query")
		}
		if !strings.Contains(fullQuery, `field2:field(name:"text") { value }`) {
			t.Errorf("Expected field2:field(name:\"text\") { value } in query")
		}
		if !strings.Contains(fullQuery, `field3:field(name:"image") { value }`) {
			t.Errorf("Expected field3:field(name:\"image\") { value } in query")
		}
	})

	t.Run("Single field", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.AddField("title", nil)

		fullQuery := builder.Build()
		if !strings.Contains(fullQuery, `field1:field(name:"title") { value }`) {
			t.Errorf("Expected field1:field(name:\"title\") { value } in query")
		}
	})

	t.Run("No fields", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetPath("/test")

		fullQuery := builder.Build()
		// Should not contain any field definitions
		if strings.Contains(fullQuery, "field1:") || strings.Contains(fullQuery, "field(name:") {
			t.Errorf("Expected no fields in query when no fields added")
		}
	})
}

// TestGetItemQueryBuilderWhereClause tests the GetItemQueryBuilder where clause building
func TestGetItemQueryBuilderWhereClause(t *testing.T) {
	t.Run("Path without existingVersionOnly", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetPath("/sitecore/content/asmblii/home")
		query := builder.Build()

		if !strings.Contains(query, `{path: "/sitecore/content/asmblii/home"}`) {
			t.Errorf("Expected where clause {path: \"/sitecore/content/asmblii/home\"} in query")
		}
	})

	t.Run("Path with existingVersionOnly=true", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetPath("/sitecore/content/asmblii/home")
		builder.SetExistingVersionOnly(true)
		query := builder.Build()

		if !strings.Contains(query, `{path: "/sitecore/content/asmblii/home", existingVersionOnly: true}`) {
			t.Errorf("Expected where clause {path: \"/sitecore/content/asmblii/home\", existingVersionOnly: true} in query")
		}
	})

	t.Run("ID without existingVersionOnly", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetItemID("{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}")
		query := builder.Build()

		if !strings.Contains(query, `{itemId: "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}"}`) {
			t.Errorf("Expected where clause {itemId: \"{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}\"} in query")
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

// TestGetItemQueryBuilderComplete tests the complete GetItemQueryBuilder
func TestGetItemQueryBuilderComplete(t *testing.T) {
	t.Run("Item by Path with fields", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetPath("/sitecore/content/asmblii/home")
		builder.AddField("title", nil)
		builder.AddField("text", nil)
		actualQuery := builder.Build()

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
			}
		}
	`

		assert.Equal(t, normalize(expectedQuery), normalize(actualQuery))
	})

	t.Run("Item by Path with existingVersionOnly", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetPath("/sitecore/content/asmblii/home")
		builder.SetExistingVersionOnly(true)
		builder.AddField("title", nil)
		actualQuery := builder.Build()

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
			}
		}
	`

		assert.Equal(t, normalize(expectedQuery), normalize(actualQuery))
	})

	t.Run("Item by ID without fields", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetItemID("{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}")
		builder.AddField("description", nil)
		actualQuery := builder.Build()

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
			}
		}
	`

		assert.Equal(t, normalize(expectedQuery), normalize(actualQuery))
	})
}

func TestParseGetItemGraphQLResponse(t *testing.T) {
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

func TestConvertFromGraphQLItemWithFields(t *testing.T) {
	t.Run("Convert GraphQL response with fields.nodes to Item", func(t *testing.T) {
		rawJSON := `{
			"itemId": "a58aab49fe074gt5b03f927c581e74d7",
			"name": "Sitecore Authoring and Management API",
			"path": "/sitecore/content/Home/Sitecore Authoring and Management API",
			"displayName": "Sitecore Authoring and Management API",
			"template": {
				"templateId": "76036f5ecbce46d1af0a4143f9b557aa",
				"name": "Sample Item"
			},
			"fields": {
				"nodes": [
					{
						"name": "Text",
						"value": "Welcome to Sitecore"
					},
					{
						"name": "Title",
						"value": "Welcome to Sitecore"
					}
				]
			}
		}`

		var graphQLItem graphQLItemWithFieldsResponse
		if err := json.Unmarshal([]byte(rawJSON), &graphQLItem); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		item := convertFromGraphQLItemWithFields(&graphQLItem)

		assert.Equal(t, "a58aab49fe074gt5b03f927c581e74d7", item.ItemID)
		assert.Equal(t, "Sitecore Authoring and Management API", item.Name)
		assert.Equal(t, "/sitecore/content/Home/Sitecore Authoring and Management API", item.Path)
		assert.Equal(t, "76036f5ecbce46d1af0a4143f9b557aa", item.TemplateID)
		assert.Equal(t, "Sample Item", item.TemplateName)

		// Check fields
		assert.Equal(t, 2, len(item.Fields))
		assert.Equal(t, "Welcome to Sitecore", item.Fields["Text"])
		assert.Equal(t, "Welcome to Sitecore", item.Fields["Title"])
	})

	t.Run("Convert GraphQL response with empty fields", func(t *testing.T) {
		rawJSON := `{
			"itemId": "test-id",
			"name": "Test Item",
			"path": "/test/path",
			"displayName": "Test Item",
			"template": {
				"templateId": "template-id",
				"name": "Test Template"
			},
			"fields": {
				"nodes": []
			}
		}`

		var graphQLItem graphQLItemWithFieldsResponse
		if err := json.Unmarshal([]byte(rawJSON), &graphQLItem); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		item := convertFromGraphQLItemWithFields(&graphQLItem)

		assert.Equal(t, "test-id", item.ItemID)
		assert.Equal(t, "Test Item", item.Name)
		assert.Equal(t, 0, len(item.Fields))
	})
}

// func TestCreateItemWithMockedClient(t *testing.T) {
// 	t.Run("Create item", func(t *testing.T) {

// 		os.Setenv("HTTPS_PROXY", "http://192.168.1.101:8081")

// 		name := "Dummy"
// 		template := "123"
// 		parentId := "234"
// 		language := "en"

// 		sampleResponse := `{
// 		}`

// 		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			assert.Equal(t, "/sitecore/api/authoring/graphql/v1/", r.URL.Path)
// 			assert.Equal(t, http.MethodPost, r.Method)

// 			w.Header().Set("Content-Type", "application/json")
// 			w.WriteHeader(http.StatusOK)
// 			_, err := w.Write([]byte(sampleResponse))
// 			if err != nil {
// 				t.Fatalf("TestListTokens write failed: %v", err)
// 			}
// 		}))
// 		defer server.Close()

// 		client := &Client{
// 			BaseURL:    server.URL,
// 			Token:      "test-token",
// 			HTTPClient: server.Client(),
// 		}

// 		fields := map[string]interface{}{}

// 		result, err := client.CreateItem(name, template, parentId, language, fields)
// 		assert.Nil(t, err)

// 		assert.NotNil(t, result)
// 	})
// }
