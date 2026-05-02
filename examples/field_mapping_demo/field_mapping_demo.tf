# Field Mapping Demo - Shows how field mapping works in practice
# This example demonstrates that fields can be accessed by their original names
# even though the API returns them as field1, field2, etc.

# Example 1: Individual item with field mapping
data "sitecoreauthoring_item" "demo_item" {
  item_id     = "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}"
  field_names = ["title", "description"]
}

# Example 2: Children with field mapping
data "sitecoreauthoring_items" "demo_children" {
  path        = "/sitecore/content"
  field_names = ["title", "description"]
}

# Output 1: Show individual item fields
output "individual_item_fields" {
  value = {
    item_id = data.sitecoreauthoring_item.demo_item.item.item_id
    name    = data.sitecoreauthoring_item.demo_item.item.name

    # ✅ Fields can be accessed by their original names
    title       = data.sitecoreauthoring_item.demo_item.item.fields["title"]
    description = data.sitecoreauthoring_item.demo_item.item.fields["description"]

    # This shows the complete fields map
    all_fields = data.sitecoreauthoring_item.demo_item.item.fields
  }
}

# Output 2: Show children fields
output "children_fields" {
  value = [for child in data.sitecoreauthoring_items.demo_children.items : {
    item_id = child.item_id
    name    = child.name

    # ✅ Children fields can also be accessed by original names
    title       = child.fields["title"]
    description = child.fields["description"]

    # This shows the complete fields map for each child
    all_fields = child.fields
  }]
}

# Output 3: Field access demonstration
output "field_access_demo" {
  value = "✅ Field mapping is working! Fields can be accessed by their original names."
}

# Output 4: Field count verification
output "field_count_verification" {
  value = length(data.sitecoreauthoring_item.demo_item.item.fields)
}