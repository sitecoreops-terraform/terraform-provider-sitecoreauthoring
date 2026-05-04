package apiclient

import (
	"fmt"
)

// GetItemByPath fetches an item by its path
func (c *Client) GetItemByPath(path string, fieldNames []string, existingVersionOnly *bool) (*Item, error) {
	// Build query using the new builder
	builder := NewGetItemQueryBuilder()
	builder.SetPath(path)
	if existingVersionOnly != nil {
		builder.SetExistingVersionOnly(*existingVersionOnly)
	}
	for _, fieldName := range fieldNames {
		builder.AddField(fieldName, nil)
	}
	query := builder.Build()

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
	// Build query using the new builder
	builder := NewGetItemQueryBuilder()
	builder.SetItemID(itemID)
	if existingVersionOnly != nil {
		builder.SetExistingVersionOnly(*existingVersionOnly)
	}
	for _, fieldName := range fieldNames {
		builder.AddField(fieldName, nil)
	}
	query := builder.Build()

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

// GetItemByIDWithFields fetches an item by its ID using the fields.nodes format
func (c *Client) GetItemByIDWithFields(itemID string, existingVersionOnly *bool) (*Item, error) {
	// Build query using the new builder - this uses fields.nodes format
	builder := NewGetItemQueryBuilder()
	builder.SetItemID(itemID)
	if existingVersionOnly != nil {
		builder.SetExistingVersionOnly(*existingVersionOnly)
	}
	// For fields.nodes format, we don't add specific fields - it gets all fields
	query := builder.Build()

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute item by ID query: %v", err)
	}

	// Parse the response using intermediate GraphQL structure
	var graphQLResponse struct {
		Item *graphQLItemWithFieldsResponse `json:"item"`
	}
	if err := parseGraphQLResponse(response.Data, &graphQLResponse); err != nil {
		return nil, fmt.Errorf("failed to parse item response: %v", err)
	}

	if graphQLResponse.Item == nil {
		return nil, fmt.Errorf("item with ID '%s' not found", itemID)
	}

	// Convert from GraphQL structure to Item structure
	item := convertFromGraphQLItemWithFields(graphQLResponse.Item)

	return item, nil
}

// GetItemByPathWithFields fetches an item by its path using the fields.nodes format
func (c *Client) GetItemByPathWithFields(path string, existingVersionOnly *bool) (*Item, error) {
	// Build query using the new builder - this uses fields.nodes format
	builder := NewGetItemQueryBuilder()
	builder.SetPath(path)
	if existingVersionOnly != nil {
		builder.SetExistingVersionOnly(*existingVersionOnly)
	}
	// For fields.nodes format, we don't add specific fields - it gets all fields
	query := builder.Build()

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute item by path query: %v", err)
	}

	// Parse the response using intermediate GraphQL structure
	var graphQLResponse struct {
		Item *graphQLItemWithFieldsResponse `json:"item"`
	}
	if err := parseGraphQLResponse(response.Data, &graphQLResponse); err != nil {
		return nil, fmt.Errorf("failed to parse item response: %v", err)
	}

	if graphQLResponse.Item == nil {
		return nil, fmt.Errorf("item with path '%s' not found", path)
	}

	// Convert from GraphQL structure to Item structure
	item := convertFromGraphQLItemWithFields(graphQLResponse.Item)

	return item, nil
}

// CreateItem creates a new item
func (c *Client) CreateItem(name string, templateID string, parentID string, language string, fields map[string]interface{}) (*Item, error) {
	// Build mutation using the new builder
	builder := NewCreateItemQueryBuilder()
	builder.SetName(name)
	builder.SetTemplateID(templateID)
	builder.SetParentID(parentID)
	builder.SetLanguage(language)
	for fieldName, fieldValue := range fields {
		builder.AddField(fieldName, fieldValue)
	}
	query := builder.Build()

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create item mutation: %v", err)
	}

	// Parse the response
	var createResponse struct {
		CreateItem struct {
			Item struct {
				ItemID string `json:"itemId"`
				Name   string `json:"name"`
				Path   string `json:"path"`
				Fields struct {
					Nodes []struct {
						Name  string      `json:"name"`
						Value interface{} `json:"value"`
					} `json:"nodes"`
				} `json:"fields"`
			} `json:"item"`
		} `json:"createItem"`
	}

	if err := parseGraphQLResponse(response.Data, &createResponse); err != nil {
		return nil, fmt.Errorf("failed to parse create item response: %v", err)
	}

	// Now fetch the full item details using the new path
	createdItem, err := c.GetItemByPathWithFields(createResponse.CreateItem.Item.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created item details: %v", err)
	}

	return createdItem, nil
}

// UpdateItem updates an existing item
func (c *Client) UpdateItem(itemID string, language string, fields map[string]interface{}, database string, path string) (*Item, error) {
	// Build mutation using the new builder
	builder := NewUpdateItemQueryBuilder()
	builder.SetItemID(itemID)
	builder.SetLanguage(language)
	for fieldName, fieldValue := range fields {
		builder.AddField(fieldName, fieldValue)
	}
	if database != "" {
		builder.SetDatabase(database)
	}
	if path != "" {
		builder.SetPath(path)
	}
	query := builder.Build()

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute update item mutation: %v", err)
	}

	// Parse the response
	var updateResponse struct {
		UpdateItem struct {
			Item struct {
				ItemID string `json:"itemId"`
				Name   string `json:"name"`
				Path   string `json:"path"`
				Fields struct {
					Nodes []struct {
						Name  string      `json:"name"`
						Value interface{} `json:"value"`
					} `json:"nodes"`
				} `json:"fields"`
			} `json:"item"`
		} `json:"updateItem"`
	}

	if err := parseGraphQLResponse(response.Data, &updateResponse); err != nil {
		return nil, fmt.Errorf("failed to parse update item response: %v", err)
	}

	// Convert fields from nodes format
	fieldsMap := make(map[string]interface{})
	for _, field := range updateResponse.UpdateItem.Item.Fields.Nodes {
		fieldsMap[field.Name] = field.Value
	}

	item := &Item{
		ItemID: updateResponse.UpdateItem.Item.ItemID,
		Name:   updateResponse.UpdateItem.Item.Name,
		Path:   updateResponse.UpdateItem.Item.Path,
		Fields: fieldsMap,
	}

	return item, nil
}

// DeleteItem deletes an item
func (c *Client) DeleteItem(path string, permanently bool) (bool, error) {
	// Build mutation using the new builder
	builder := NewDeleteItemQueryBuilder()
	builder.SetPath(path)
	builder.SetPermanently(permanently)
	query := builder.Build()

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return false, fmt.Errorf("failed to execute delete item mutation: %v", err)
	}

	// Parse the response
	var deleteResponse struct {
		DeleteItem struct {
			Successful bool `json:"successful"`
		} `json:"deleteItem"`
	}

	if err := parseGraphQLResponse(response.Data, &deleteResponse); err != nil {
		return false, fmt.Errorf("failed to parse delete item response: %v", err)
	}

	return deleteResponse.DeleteItem.Successful, nil
}

// graphQLItemWithFieldsResponse represents the intermediate structure for GraphQL responses with fields.nodes format
type graphQLItemWithFieldsResponse struct {
	ItemID      string `json:"itemId"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Template    struct {
		TemplateID string `json:"templateId"`
		Name       string `json:"name"`
	} `json:"template"`
	Fields struct {
		Nodes []struct {
			Name  string      `json:"name"`
			Value interface{} `json:"value"`
		} `json:"nodes"`
	} `json:"fields"`
}

// convertFromGraphQLItemWithFields converts from the GraphQL response structure with fields.nodes to our Item structure
func convertFromGraphQLItemWithFields(graphQLItem *graphQLItemWithFieldsResponse) *Item {
	if graphQLItem == nil {
		return nil
	}

	// Convert fields from nodes format
	fieldsMap := make(map[string]interface{})
	for _, field := range graphQLItem.Fields.Nodes {
		fieldsMap[field.Name] = field.Value
	}

	return &Item{
		ItemID:       graphQLItem.ItemID,
		Path:         graphQLItem.Path,
		Name:         graphQLItem.Name,
		DisplayName:  graphQLItem.DisplayName,
		TemplateID:   graphQLItem.Template.TemplateID,
		TemplateName: graphQLItem.Template.Name,
		Fields:       fieldsMap,
	}
}
