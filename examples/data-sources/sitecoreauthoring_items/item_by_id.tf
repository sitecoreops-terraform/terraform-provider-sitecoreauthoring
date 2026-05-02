# Example: Get children of an item by ID with specific fields
data "sitecoreauthoring_items" "by_id" {
  item_id     = "110d559fdea542ea9c1c8a5df7e70ef9"
  field_names = ["Title", "Name"]
}


# Output children information
output "children_by_id" {
  value = [for child in data.sitecoreauthoring_items.by_id.items : {
    item_id       = child.item_id
    name          = child.name
    display_name  = child.display_name
    template_name = child.template_name
    fields        = child.fields
    title         = child.fields["Title"]
  }]
}

output "children_count" {
  value = length(data.sitecoreauthoring_items.by_id.items)
}
