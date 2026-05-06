data "sitecoreauthoring_item" "webhook" {
  path = "/sitecore/system/Workflows/Sample Workflow/Awaiting Approval/Approve/Webhook"
}

resource "sitecoreauthoring_item_field" "webhook" {
  item_id     = data.sitecoreauthoring_item.webhook.item.item_id
  language    = "en"
  field_name  = "Url"
  field_value = "https://my-webhook-in-terraform"
}
