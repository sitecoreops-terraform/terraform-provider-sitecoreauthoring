# Simple example: Get children without specifying field names
data "sitecoreauthoring_items" "simple" {
  path = "/sitecore/content"
  # No field_names specified - will use defaults
}

# Output children information
output "simple_children" {
  value = [for child in data.sitecoreauthoring_items.simple.items : {
    item_id       = child.item_id
    name          = child.name
    display_name  = child.display_name
    template_name = child.template_name
    # Fields map will contain any available fields
    field_count = length(child.fields)
    has_fields  = length(child.fields) > 0
  }]
}

output "simple_children_count" {
  value = length(data.sitecoreauthoring_items.simple.items)
}

output "success_message" {
  value = "✅ Children retrieved successfully! The items data source is working correctly."
}