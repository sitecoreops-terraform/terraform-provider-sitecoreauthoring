# Sitecore Item Data Source

The `sitecoreauthoring_item` data source allows you to retrieve information about a Sitecore item by either its ID or path from the Sitecore Authoring API.

## Example Usage

### By Path

```hcl
data "sitecoreauthoring_item" "by_path" {
  path = "/sitecore/content/asmblii/home"
  field_names = ["title", "text", "image"]
  existing_version_only = true
}

output "item_info" {
  value = {
    item_id      = data.sitecoreauthoring_item.by_path.item.item_id
    path         = data.sitecoreauthoring_item.by_path.item.path
    name         = data.sitecoreauthoring_item.by_path.item.name
    display_name = data.sitecoreauthoring_item.by_path.item.display_name
    template_id  = data.sitecoreauthoring_item.by_path.item.template_id
    template_name = data.sitecoreauthoring_item.by_path.item.template_name
    fields       = data.sitecoreauthoring_item.by_path.item.fields
  }
}
```

### By ID

```hcl
data "sitecoreauthoring_item" "by_id" {
  item_id = "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}"
  field_names = ["title", "description"]
}

output "item_info" {
  value = {
    item_id      = data.sitecoreauthoring_item.by_id.item.item_id
    path         = data.sitecoreauthoring_item.by_id.item.path
    name         = data.sitecoreauthoring_item.by_id.item.name
    display_name = data.sitecoreauthoring_item.by_id.item.display_name
    template_id  = data.sitecoreauthoring_item.by_id.item.template_id
    template_name = data.sitecoreauthoring_item.by_id.item.template_name
    fields       = data.sitecoreauthoring_item.by_id.item.fields
  }
}
```

## Argument Reference

- **item_id** (Optional, string) - The ID of the item to retrieve. Either `item_id` or `path` must be specified.
- **path** (Optional, string) - The path of the item to retrieve. Either `item_id` or `path` must be specified.
- **field_names** (Optional, set of strings) - Set of field names to retrieve for the item and its children. If not specified, no fields will be retrieved.
- **existing_version_only** (Optional, bool) - If true, only returns items that have an existing version. If not specified, returns items regardless of version status.

## Attributes Reference

- **id** - The identifier for the data source, set to the item ID or path.
- **item** - The retrieved item with the following attributes:
  - **item_id** (string) - The ID of the item.
  - **path** (string) - The path of the item.
  - **name** (string) - The name of the item.
  - **display_name** (string) - The display name of the item.
  - **template_id** (string) - The template ID of the item.
  - **template_name** (string) - The template name of the item.
  - **fields** (map) - The requested fields of the item as key-value pairs.
  - **children** (list) - The direct children of the item, each with the same structure as the parent item.

## GraphQL Query Structure

The data source uses the Sitecore Authoring API GraphQL endpoint with queries structured as:

```graphql
query {
  item(where: {
    itemId: "{GUID}",  # or path: "/path/to/item"
    existingVersionOnly: true/false  # optional
  }) {
    itemId
    path
    name
    displayName
    templateId
    templateName
    fields {
      field1
      field2
      # ... requested fields
    }
    children {
      itemId
      path
      name
      displayName
      templateId
      templateName
      fields {
        field1
        field2
        # ... requested fields
      }
    }
  }
}
```

## Notes

- Exactly one of `item_id` or `path` must be specified.
- If `existing_version_only` is not specified, it will be omitted from the query (default behavior).
- The data source returns an empty collection instead of null when no items match the criteria.
- Children are automatically retrieved with the same field structure as the parent item.