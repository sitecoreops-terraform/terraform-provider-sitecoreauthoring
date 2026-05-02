# Example: Get children of the content item (which should have tenants)
data "sitecoreauthoring_items" "content_children" {
  path        = "/sitecore/content"
  field_names = ["title", "tenantName"]
}

# Output tenants information
output "content_children" {
  value = [for child in data.sitecoreauthoring_items.content_children.items : {
    item_id       = child.item_id
    name          = child.name
    display_name  = child.display_name
    template_name = child.template_name
    fields        = child.fields
  }]
}

output "content_children_count" {
  value = length(data.sitecoreauthoring_items.content_children.items)
}