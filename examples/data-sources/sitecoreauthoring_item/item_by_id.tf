# Example: Get item by ID with specific fields
data "sitecoreauthoring_item" "by_id" {
  item_id     = "{110D559F-DEA5-42EA-9C1C-8A5DF7E70EF9}"
  field_names = ["Title", "Text"]
}

# Output item information
output "item_by_id" {
  value = {
    item_id       = data.sitecoreauthoring_item.by_id.item.item_id
    path          = data.sitecoreauthoring_item.by_id.item.path
    name          = data.sitecoreauthoring_item.by_id.item.name
    display_name  = data.sitecoreauthoring_item.by_id.item.display_name
    template_id   = data.sitecoreauthoring_item.by_id.item.template_id
    template_name = data.sitecoreauthoring_item.by_id.item.template_name
    fields        = data.sitecoreauthoring_item.by_id.item.fields
    title         = data.sitecoreauthoring_item.by_id.item.fields["Title"]
  }
}
