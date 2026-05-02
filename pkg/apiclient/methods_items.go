package apiclient

import (
	"fmt"
)

// buildFieldsQuery builds the fields query string from field names
func buildFieldsQuery(fieldNames []string) string {
	if len(fieldNames) == 0 {
		return ""
	}
	var fieldsQuery string
	for i, fieldName := range fieldNames {
		// Create field alias like field1:field(name:"Title") { value }
		fieldAlias := fmt.Sprintf(`field%d`, i+1)
		fieldsQuery += fmt.Sprintf(`%s:field(name:"%s") { value }
					`, fieldAlias, fieldName)
	}
	return fieldsQuery
}

// buildWhereClause builds the where clause for item queries
func buildWhereClause(usePath bool, identifier string, existingVersionOnly *bool) string {
	var whereClause string
	if usePath {
		whereClause = fmt.Sprintf(`{path: "%s"`, identifier)
	} else {
		whereClause = fmt.Sprintf(`{itemId: "%s"`, identifier)
	}
	if existingVersionOnly != nil {
		whereClause += fmt.Sprintf(`, existingVersionOnly: %v`, *existingVersionOnly)
	}
	whereClause += "}"
	return whereClause
}

// buildItemQuery builds the complete GraphQL query for item retrieval
func buildItemQuery(whereClause string, fieldsQuery string) string {
	return fmt.Sprintf(`
		query ItemLookup {
			item(where: %s) {
				itemId
				path
				name
				displayName
				template {
					templateId
					name
				}
				%s
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
						%s
					}
				}
			}
		}
	`, whereClause, fieldsQuery, fieldsQuery)
}

// GetItemByPath fetches an item by its path
func (c *Client) GetItemByPath(path string, fieldNames []string, existingVersionOnly *bool) (*Item, error) {
	fieldsQuery := buildFieldsQuery(fieldNames)
	whereClause := buildWhereClause(true, path, existingVersionOnly)
	query := buildItemQuery(whereClause, fieldsQuery)

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute item by path query: %v", err)
	}

	// Parse the response using intermediate GraphQL structure
	var graphQLResponse struct {
		Item *graphQLItemResponse `json:"item"`
	}
	if err := parseGraphQLResponse(response.Data, &graphQLResponse); err != nil {
		return nil, fmt.Errorf("failed to parse item response: %v", err)
	}

	if graphQLResponse.Item == nil {
		return nil, fmt.Errorf("item with path '%s' not found", path)
	}

	// Convert from GraphQL structure to Item structure
	item := convertFromGraphQLItem(graphQLResponse.Item, fieldNames)

	return item, nil
}

// GetItemByID fetches an item by its ID
func (c *Client) GetItemByID(itemID string, fieldNames []string, existingVersionOnly *bool) (*Item, error) {
	fieldsQuery := buildFieldsQuery(fieldNames)
	whereClause := buildWhereClause(false, itemID, existingVersionOnly)
	query := buildItemQuery(whereClause, fieldsQuery)

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute item by ID query: %v", err)
	}

	// Parse the response using intermediate GraphQL structure
	var graphQLResponse struct {
		Item *graphQLItemResponse `json:"item"`
	}
	if err := parseGraphQLResponse(response.Data, &graphQLResponse); err != nil {
		return nil, fmt.Errorf("failed to parse item response: %v", err)
	}

	if graphQLResponse.Item == nil {
		return nil, fmt.Errorf("item with ID '%s' not found", itemID)
	}

	// Convert from GraphQL structure to Item structure
	item := convertFromGraphQLItem(graphQLResponse.Item, fieldNames)

	return item, nil
}

// convertFromGraphQLItem converts from the GraphQL response structure to our Item structure
func convertFromGraphQLItem(graphQLItem *graphQLItemResponse, fieldNames []string) *Item {
	if graphQLItem == nil {
		return nil
	}

	// Convert DynamicFields to the format expected by TransformGraphQLFields
	rawFields := make(map[string]interface{})
	for alias, fieldData := range graphQLItem.DynamicFields {
		rawFields[alias] = fieldData
	}

	// Convert fields using TransformGraphQLFields
	transformedFields := TransformGraphQLFields(rawFields)

	// Map field aliases back to original field names
	// Create mapping: field1 -> originalFieldName1, field2 -> originalFieldName2, etc.
	finalFields := make(map[string]interface{})
	if len(fieldNames) > 0 {
		for i, originalFieldName := range fieldNames {
			fieldAlias := fmt.Sprintf("field%d", i+1)
			if value, exists := transformedFields[fieldAlias]; exists {
				finalFields[originalFieldName] = value
			}
		}
	} else {
		// If no field names provided, use the transformed fields as-is
		finalFields = transformedFields
	}

	// Convert children
	var children []Item
	for _, child := range graphQLItem.Children.Nodes {
		// For children, we don't have the original field names, so use empty slice
		// The children will have the same field structure as parsed from GraphQL
		convertedChild := convertFromGraphQLItem(&child, nil)
		if convertedChild != nil {
			children = append(children, *convertedChild)
		}
	}

	return &Item{
		ItemID:       graphQLItem.ItemID,
		Path:         graphQLItem.Path,
		Name:         graphQLItem.Name,
		DisplayName:  graphQLItem.DisplayName,
		TemplateID:   graphQLItem.Template.TemplateID,
		TemplateName: graphQLItem.Template.Name,
		Fields:       finalFields,
		Children:     children,
	}
}
