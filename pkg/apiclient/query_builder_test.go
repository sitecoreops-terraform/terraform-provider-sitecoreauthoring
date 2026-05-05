package apiclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetItemQueryBuilder tests the GetItemQueryBuilder
func TestGetItemQueryBuilder(t *testing.T) {
	t.Run("Build query with path and fields", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetPath("/sitecore/content/home")
		builder.AddField("title", nil)
		builder.AddField("text", nil)

		query := builder.Build()

		expectedQuery := `
		query ItemLookup {
			item(where: {path: "/sitecore/content/home"}) {
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

		assert.Equal(t, normalize(expectedQuery), normalize(query))
	})

	t.Run("Build query with item ID and existingVersionOnly", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetItemID("{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}")
		builder.SetExistingVersionOnly(true)
		builder.AddField("description", nil)

		query := builder.Build()

		expectedQuery := `
		query ItemLookup {
			item(where: {itemId: "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}", existingVersionOnly: true}) {
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

		assert.Equal(t, normalize(expectedQuery), normalize(query))
	})

	t.Run("Build query with no fields", func(t *testing.T) {
		builder := NewGetItemQueryBuilder()
		builder.SetPath("/sitecore/content/test")

		query := builder.Build()

		expectedQuery := `
		query ItemLookup {
			item(where: {path: "/sitecore/content/test"}) {
				itemId
				path
				name
				displayName
				template {
					templateId
					name
				}
				
			}
		}
	`

		assert.Equal(t, expectedQuery, query)
	})
}

// TestCreateItemQueryBuilder tests the CreateItemQueryBuilder
func TestCreateItemQueryBuilder(t *testing.T) {
	t.Run("Build create mutation with fields", func(t *testing.T) {
		builder := NewCreateItemQueryBuilder()
		builder.SetName("Test Item")
		builder.SetTemplateID("{76036F5E-CBCE-46D1-AF0A-4143F9B557AA}")
		builder.SetParentID("{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}")
		builder.SetLanguage("en")
		builder.AddField("title", "Welcome to Sitecore")
		builder.AddField("text", "Welcome to Sitecore")

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			createItem(
				input: {
					name: "Test Item"
					templateId: "{76036F5E-CBCE-46D1-AF0A-4143F9B557AA}"
					parent: "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}"
					language: "en"
					fields: [{name: "text", value: "Welcome to Sitecore"}, {name: "title", value: "Welcome to Sitecore"}]
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

		assert.Equal(t, expectedMutation, mutation)
	})

	t.Run("Build create mutation without fields", func(t *testing.T) {
		builder := NewCreateItemQueryBuilder()
		builder.SetName("Empty Item")
		builder.SetTemplateID("{TEMPLATE-ID}")
		builder.SetParentID("{PARENT-ID}")
		builder.SetLanguage("en")

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			createItem(
				input: {
					name: "Empty Item"
					templateId: "{TEMPLATE-ID}"
					parent: "{PARENT-ID}"
					language: "en"
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

		assert.Equal(t, expectedMutation, mutation)
	})

	t.Run("Build create mutation with null field", func(t *testing.T) {
		builder := NewCreateItemQueryBuilder()
		builder.SetName("Item with null")
		builder.SetTemplateID("{TEMPLATE-ID}")
		builder.SetParentID("{PARENT-ID}")
		builder.SetLanguage("en")
		builder.AddField("title", "Test Title")
		builder.AddField("empty", nil)

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			createItem(
				input: {
					name: "Item with null"
					templateId: "{TEMPLATE-ID}"
					parent: "{PARENT-ID}"
					language: "en"
					fields: [{name: "empty", value: null}, {name: "title", value: "Test Title"}]
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

		assert.Equal(t, expectedMutation, mutation)
	})
}

// TestUpdateItemQueryBuilder tests the UpdateItemQueryBuilder
func TestUpdateItemQueryBuilder(t *testing.T) {
	t.Run("Build update mutation with fields", func(t *testing.T) {
		builder := NewUpdateItemQueryBuilder()
		builder.SetItemID("{59C9BA60-6483-451C-A435-B60BED2DBA75}")
		builder.SetLanguage("en")
		builder.AddField("Title", "My new page")
		builder.AddField("Content", "Lorem Ipsum")

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			updateItem(
				input: {
					itemId: "{59C9BA60-6483-451C-A435-B60BED2DBA75}"
					language: "en"
					fields: [
					{name: "Title", value: "My new page", reset: false}
					,
					{name: "Content", value: "Lorem Ipsum", reset: false}
					]
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

		assert.Equal(t, normalize(expectedMutation), normalize(mutation))
	})

	t.Run("Build update mutation with database and path", func(t *testing.T) {
		builder := NewUpdateItemQueryBuilder()
		builder.SetItemID("{ITEM-ID}")
		builder.SetLanguage("en")
		builder.AddField("Title", "Updated Title")
		builder.SetDatabase("master")
		builder.SetPath("/sitecore/content/mycollection/mysite/Home/PageTest")

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			updateItem(
				input: {
					itemId: "{ITEM-ID}"
					language: "en"
					fields: [{name: "Title", value: "Updated Title", reset: false}]
					database: "master"
					path: "/sitecore/content/mycollection/mysite/Home/PageTest"
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

		assert.Equal(t, expectedMutation, mutation)
	})
}

// TestDeleteItemQueryBuilder tests the DeleteItemQueryBuilder
func TestDeleteItemQueryBuilder(t *testing.T) {
	t.Run("Build delete mutation permanently", func(t *testing.T) {
		builder := NewDeleteItemQueryBuilder()
		builder.SetPath("/sitecore/content/Home/Test")
		builder.SetPermanently(true)

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			deleteItem(
				input: {
					path: "/sitecore/content/Home/Test"
					permanently: true
				}
			) {
				successful
			}
		}
	`

		assert.Equal(t, expectedMutation, mutation)
	})

	t.Run("Build delete mutation to recycle bin", func(t *testing.T) {
		builder := NewDeleteItemQueryBuilder()
		builder.SetPath("/sitecore/content/Test")
		builder.SetPermanently(false)

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			deleteItem(
				input: {
					path: "/sitecore/content/Test"
					permanently: false
				}
			) {
				successful
			}
		}
	`

		assert.Equal(t, expectedMutation, mutation)
	})
}

// TestGetChildItemsQueryBuilder tests the GetChildItemsQueryBuilder
func TestGetChildItemsQueryBuilder(t *testing.T) {
	t.Run("Build child items query", func(t *testing.T) {
		builder := NewGetChildItemsQueryBuilder()
		builder.SetPath("/sitecore/content/parent")
		builder.AddField("title", nil)

		query := builder.Build()

		expectedQuery := `
		query ItemLookup {
			item(where: {path: "/sitecore/content/parent"}) {
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
						field1:field(name:"title") { value }
						
					}
				}
			}
		}
	`

		assert.Equal(t, normalize(expectedQuery), normalize(query))
	})
}

// TestRenameItemQueryBuilder tests the RenameItemQueryBuilder
func TestRenameItemQueryBuilder(t *testing.T) {
	t.Run("Build rename mutation with item ID and new name", func(t *testing.T) {
		builder := NewRenameItemQueryBuilder()
		builder.SetItemID("{60FD672A-D787-4109-B823-3DB1A45DB4E4}")
		builder.SetNewName("sub2")

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			renameItem(
				input: {
					itemId: "{60FD672A-D787-4109-B823-3DB1A45DB4E4}"
					newName: "sub2"
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

		assert.Equal(t, expectedMutation, mutation)
	})

	t.Run("Build rename mutation with database", func(t *testing.T) {
		builder := NewRenameItemQueryBuilder()
		builder.SetItemID("{60FD672A-D787-4109-B823-3DB1A45DB4E4}")
		builder.SetNewName("sub2")
		builder.SetDatabase("master")

		mutation := builder.Build()

		expectedMutation := `
		mutation {
			renameItem(
				input: {
					itemId: "{60FD672A-D787-4109-B823-3DB1A45DB4E4}"
					newName: "sub2"
					database: "master"
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

		assert.Equal(t, expectedMutation, mutation)
	})
}
