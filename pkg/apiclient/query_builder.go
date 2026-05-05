package apiclient

import (
	"fmt"
	"sort"
	"strings"
)

// QueryBuilder is the base class for building GraphQL queries
type QueryBuilder interface {
	Build() string
}

// BaseQueryBuilder provides common functionality for all query builders
type BaseQueryBuilder struct {
	fields              map[string]interface{}
	fieldNames          []string
	existingVersionOnly *bool
}

// NewBaseQueryBuilder creates a new BaseQueryBuilder
func NewBaseQueryBuilder() *BaseQueryBuilder {
	return &BaseQueryBuilder{
		fields:              make(map[string]interface{}),
		fieldNames:          make([]string, 0),
		existingVersionOnly: nil,
	}
}

// AddField adds a field to the query
func (b *BaseQueryBuilder) AddField(name string, value interface{}) *BaseQueryBuilder {
	b.fields[name] = value
	b.fieldNames = append(b.fieldNames, name)
	return b
}

// SetExistingVersionOnly sets the existingVersionOnly flag
func (b *BaseQueryBuilder) SetExistingVersionOnly(existingVersionOnly bool) *BaseQueryBuilder {
	b.existingVersionOnly = &existingVersionOnly
	return b
}

// buildFieldsQuery builds the fields query string from field names
func (b *BaseQueryBuilder) buildFieldsQuery() string {
	if len(b.fieldNames) == 0 {
		return ""
	}
	var fieldsQuery string
	for i, fieldName := range b.fieldNames {
		// Create field alias like field1:field(name:"Title") { value }
		fieldAlias := fmt.Sprintf(`field%d`, i+1)
		fieldsQuery += fmt.Sprintf(`%s:field(name:"%s") { value }`, fieldAlias, fieldName)
		if i < len(b.fieldNames)-1 {
			fieldsQuery += "\n\t\t\t\t\t"
		} else {
			fieldsQuery += "\n\t\t\t\t\t"
		}
	}
	return fieldsQuery
}

// GetItemQueryBuilder builds queries for getting items
type GetItemQueryBuilder struct {
	*BaseQueryBuilder
	identifier string
	usePath    bool
}

// NewGetItemQueryBuilder creates a new GetItemQueryBuilder
func NewGetItemQueryBuilder() *GetItemQueryBuilder {
	return &GetItemQueryBuilder{
		BaseQueryBuilder: NewBaseQueryBuilder(),
		identifier:       "",
		usePath:          true,
	}
}

// SetPath sets the item path
func (b *GetItemQueryBuilder) SetPath(path string) *GetItemQueryBuilder {
	b.identifier = path
	b.usePath = true
	return b
}

// SetItemID sets the item ID
func (b *GetItemQueryBuilder) SetItemID(itemID string) *GetItemQueryBuilder {
	b.identifier = itemID
	b.usePath = false
	return b
}

// Build builds the complete GraphQL query
func (b *GetItemQueryBuilder) Build() string {
	whereClause := b.buildWhereClause()
	fieldsQuery := b.buildFieldsQuery()
	return b.buildItemQuery(whereClause, fieldsQuery)
}

// buildWhereClause builds the where clause for item queries
func (b *GetItemQueryBuilder) buildWhereClause() string {
	var whereClause string
	if b.usePath {
		whereClause = fmt.Sprintf(`{path: "%s"`, b.identifier)
	} else {
		whereClause = fmt.Sprintf(`{itemId: "%s"`, b.identifier)
	}
	if b.existingVersionOnly != nil {
		whereClause += fmt.Sprintf(`, existingVersionOnly: %v`, *b.existingVersionOnly)
	}
	whereClause += "}"
	return whereClause
}

// buildItemQuery builds the complete GraphQL query for item retrieval
func (b *GetItemQueryBuilder) buildItemQuery(whereClause string, fieldsQuery string) string {
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
			}
		}
	`, whereClause, fieldsQuery)
}

func (b *GetItemQueryBuilder) buildItemChildrenQuery(whereClause string, fieldsQuery string) string {
	return fmt.Sprintf(`
		query ItemLookup {
			item(where: %s) {
				itemId
				path
				name
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
	`, whereClause, fieldsQuery)
}

// CreateItemQueryBuilder builds mutations for creating items
type CreateItemQueryBuilder struct {
	name       string
	templateID string
	parentID   string
	language   string
	fields     map[string]interface{}
}

// NewCreateItemQueryBuilder creates a new CreateItemQueryBuilder
func NewCreateItemQueryBuilder() *CreateItemQueryBuilder {
	return &CreateItemQueryBuilder{
		fields: make(map[string]interface{}),
	}
}

// SetName sets the item name
func (b *CreateItemQueryBuilder) SetName(name string) *CreateItemQueryBuilder {
	b.name = name
	return b
}

// SetTemplateID sets the template ID
func (b *CreateItemQueryBuilder) SetTemplateID(templateID string) *CreateItemQueryBuilder {
	b.templateID = templateID
	return b
}

// SetParentID sets the parent ID
func (b *CreateItemQueryBuilder) SetParentID(parentID string) *CreateItemQueryBuilder {
	b.parentID = parentID
	return b
}

// SetLanguage sets the language
func (b *CreateItemQueryBuilder) SetLanguage(language string) *CreateItemQueryBuilder {
	b.language = language
	return b
}

// AddField adds a field to the item
func (b *CreateItemQueryBuilder) AddField(name string, value interface{}) *CreateItemQueryBuilder {
	b.fields[name] = value
	return b
}

// Build builds the complete GraphQL mutation
func (b *CreateItemQueryBuilder) Build() string {
	fieldsQuery := ""
	if len(b.fields) > 0 {
		fieldsQuery = "fields: ["

		// Sort field names for consistent output
		fieldNames := make([]string, 0, len(b.fields))
		for fieldName := range b.fields {
			fieldNames = append(fieldNames, fieldName)
		}
		sort.Strings(fieldNames)

		for i, fieldName := range fieldNames {
			if i > 0 {
				fieldsQuery += ", "
			}
			fieldValue := b.fields[fieldName]
			// Convert field value to string
			var valueStr string
			if fieldValue == nil {
				valueStr = "null"
			} else {
				// Escape quotes and handle special characters
				escapedValue := fmt.Sprintf("%v", fieldValue)
				// Simple JSON escaping - replace " with \"
				escapedValue = strings.ReplaceAll(escapedValue, `"`, `\"`)
				valueStr = `"` + escapedValue + `"`
			}
			fieldsQuery += fmt.Sprintf(`{name: "%s", value: %s}`, fieldName, valueStr)
		}
		fieldsQuery += "]"
	}

	if fieldsQuery == "" {
		return fmt.Sprintf(`
		mutation {
			createItem(
				input: {
					name: "%s"
					templateId: "%s"
					parent: "%s"
					language: "%s"
				}
			) {
				item {
					itemId
					name
					path
					fields(ownFields: true) {
						nodes {
							name
							value
						}
					}
				}
			}
		}
	`, b.name, b.templateID, b.parentID, b.language)
	} else {
		return fmt.Sprintf(`
		mutation {
			createItem(
				input: {
					name: "%s"
					templateId: "%s"
					parent: "%s"
					language: "%s"
					%s
				}
			) {
				item {
					itemId
					name
					path
					fields(ownFields: true) {
						nodes {
							name
							value
						}
					}
				}
			}
		}
	`, b.name, b.templateID, b.parentID, b.language, fieldsQuery)
	}
}

// UpdateItemQueryBuilder builds mutations for updating items
type UpdateItemQueryBuilder struct {
	itemID   string
	language string
	fields   map[string]interface{}
	database string
	path     string
}

// NewUpdateItemQueryBuilder creates a new UpdateItemQueryBuilder
func NewUpdateItemQueryBuilder() *UpdateItemQueryBuilder {
	return &UpdateItemQueryBuilder{
		fields: make(map[string]interface{}),
	}
}

// SetItemID sets the item ID
func (b *UpdateItemQueryBuilder) SetItemID(itemID string) *UpdateItemQueryBuilder {
	b.itemID = itemID
	return b
}

// SetLanguage sets the language
func (b *UpdateItemQueryBuilder) SetLanguage(language string) *UpdateItemQueryBuilder {
	b.language = language
	return b
}

// AddField adds a field to update
func (b *UpdateItemQueryBuilder) AddField(name string, value interface{}) *UpdateItemQueryBuilder {
	b.fields[name] = value
	return b
}

// SetDatabase sets the database
func (b *UpdateItemQueryBuilder) SetDatabase(database string) *UpdateItemQueryBuilder {
	b.database = database
	return b
}

// SetPath sets the path
func (b *UpdateItemQueryBuilder) SetPath(path string) *UpdateItemQueryBuilder {
	b.path = path
	return b
}

// Build builds the complete GraphQL mutation
func (b *UpdateItemQueryBuilder) Build() string {
	fieldsQuery := ""
	if len(b.fields) > 0 {
		fieldsQuery = "fields: ["

		// Sort field names alphabetically for consistent output
		fieldNames := make([]string, 0, len(b.fields))
		for fieldName := range b.fields {
			fieldNames = append(fieldNames, fieldName)
		}
		sort.Strings(fieldNames)

		for i, fieldName := range fieldNames {
			fieldValue := b.fields[fieldName]
			if i > 0 {
				fieldsQuery += ", "
			}
			// Convert field value to string
			var valueStr string
			if fieldValue == nil {
				valueStr = "null"
			} else {
				// Escape quotes and handle special characters
				escapedValue := fmt.Sprintf("%v", fieldValue)
				// Simple JSON escaping - replace " with \
				escapedValue = strings.ReplaceAll(escapedValue, `"`, `\"`)
				valueStr = `"` + escapedValue + `"`
			}
			fieldsQuery += fmt.Sprintf(`{name: "%s", value: %s, reset: false}`, fieldName, valueStr)
		}
		fieldsQuery += "]"
	}

	// Build the mutation with optional parameters
	mutation := fmt.Sprintf(`
		mutation {
			updateItem(
				input: {
					itemId: "%s"
					language: "%s"
					%s`, b.itemID, b.language, fieldsQuery)

	if b.database != "" {
		mutation += fmt.Sprintf(`
					database: "%s"`, b.database)
	}
	if b.path != "" {
		mutation += fmt.Sprintf(`
					path: "%s"`, b.path)
	}

	mutation += `
				}
			) {
				item {
					itemId
					name
					path
					fields(ownFields: true) {
						nodes {
							name
							value
						}
					}
				}
			}
		}
	`

	return mutation
}

// DeleteItemQueryBuilder builds mutations for deleting items
type DeleteItemQueryBuilder struct {
	path        string
	permanently bool
}

// NewDeleteItemQueryBuilder creates a new DeleteItemQueryBuilder
func NewDeleteItemQueryBuilder() *DeleteItemQueryBuilder {
	return &DeleteItemQueryBuilder{}
}

// SetPath sets the item path
func (b *DeleteItemQueryBuilder) SetPath(path string) *DeleteItemQueryBuilder {
	b.path = path
	return b
}

// SetPermanently sets whether to delete permanently
func (b *DeleteItemQueryBuilder) SetPermanently(permanently bool) *DeleteItemQueryBuilder {
	b.permanently = permanently
	return b
}

// Build builds the complete GraphQL mutation
func (b *DeleteItemQueryBuilder) Build() string {
	return fmt.Sprintf(`
		mutation {
			deleteItem(
				input: {
					path: "%s"
					permanently: %v
				}
			) {
				successful
			}
		}
	`, b.path, b.permanently)
}

// RenameItemQueryBuilder builds mutations for renaming items
type RenameItemQueryBuilder struct {
	itemID   string
	newName  string
	database string
}

// NewRenameItemQueryBuilder creates a new RenameItemQueryBuilder
func NewRenameItemQueryBuilder() *RenameItemQueryBuilder {
	return &RenameItemQueryBuilder{}
}

// SetItemID sets the item ID
func (b *RenameItemQueryBuilder) SetItemID(itemID string) *RenameItemQueryBuilder {
	b.itemID = itemID
	return b
}

// SetNewName sets the new name
func (b *RenameItemQueryBuilder) SetNewName(newName string) *RenameItemQueryBuilder {
	b.newName = newName
	return b
}

// SetDatabase sets the database
func (b *RenameItemQueryBuilder) SetDatabase(database string) *RenameItemQueryBuilder {
	b.database = database
	return b
}

// Build builds the complete GraphQL mutation
func (b *RenameItemQueryBuilder) Build() string {
	mutation := fmt.Sprintf(`
		mutation {
			renameItem(
				input: {
					itemId: "%s"
					newName: "%s"`, b.itemID, b.newName)

	if b.database != "" {
		mutation += fmt.Sprintf(`
					database: "%s"`, b.database)
	}

	mutation += `
				}
			) {
				item {
					itemId
					path
					name
					fields(ownFields: true) {
						nodes {
							name
							value
						}
					}
				}
			}
		}`

	return mutation
}

// GetChildItemsQueryBuilder builds queries for getting child items
type GetChildItemsQueryBuilder struct {
	*GetItemQueryBuilder
}

// NewGetChildItemsQueryBuilder creates a new GetChildItemsQueryBuilder
func NewGetChildItemsQueryBuilder() *GetChildItemsQueryBuilder {
	return &GetChildItemsQueryBuilder{
		GetItemQueryBuilder: NewGetItemQueryBuilder(),
	}
}

// Build builds the complete GraphQL query for child items
func (b *GetChildItemsQueryBuilder) Build() string {
	// For child items, we use the same query structure but focus on the children part
	whereClause := b.buildWhereClause()
	fieldsQuery := b.buildFieldsQuery()
	return b.buildItemChildrenQuery(whereClause, fieldsQuery)
}
