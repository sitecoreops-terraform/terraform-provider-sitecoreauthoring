data "sitecoreauthoring_item" "template" {
  path = "/sitecore/templates/Common/Folder"
}

data "sitecoreauthoring_item" "parent" {
  path = "/sitecore/content"
}

output "parent_id" {
  value = data.sitecoreauthoring_item.template.item_id
}

output "template_id" {
  value = data.sitecoreauthoring_item.parent.item_id
}
resource "sitecoreauthoring_item" "setting" {
  parent_id   = data.sitecoreauthoring_item.parent.item.item_id
  template_id = data.sitecoreauthoring_item.template.item.item_id
  name        = "Demo"
  language    = "en"
}
