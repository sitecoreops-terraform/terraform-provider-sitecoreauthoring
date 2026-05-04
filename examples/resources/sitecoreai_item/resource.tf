#
# This is an example of creating an item with dynamic lookup of template and parent
#

data "sitecoreauthoring_item" "template" {
  path = "/sitecore/templates/Common/Folder"
}

data "sitecoreauthoring_item" "parent" {
  path = "/sitecore/content"
}

resource "sitecoreauthoring_item" "setting" {
  parent_id   = data.sitecoreauthoring_item.parent.item.item_id
  template_id = data.sitecoreauthoring_item.template.item.item_id
  name        = "Demo"
  language    = "en"
}

output "parent_id" {
  value = data.sitecoreauthoring_item.template.item.item_id
}

output "template_id" {
  value = data.sitecoreauthoring_item.parent.item.item_id
}

output "created_item_id" {
  value = sitecoreauthoring_item.setting.item_id
}
