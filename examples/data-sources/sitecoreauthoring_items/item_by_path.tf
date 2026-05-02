# Example: Get children of an item by path with specific fields
data "sitecoreauthoring_items" "by_path" {
  path        = "/sitecore/content/"
  field_names = ["Name", "Title"]
}

# Output children information
output "children_by_path" {
  value = [for child in data.sitecoreauthoring_items.by_path.items : {
    item_id       = child.item_id
    name          = child.name
    display_name  = child.display_name
    template_name = child.template_name
    name          = child.fields["Name"]
    title         = child.fields["Title"]
  }]
}

output "children_by_path_count" {
  value = length(data.sitecoreauthoring_items.by_path.items)
}
