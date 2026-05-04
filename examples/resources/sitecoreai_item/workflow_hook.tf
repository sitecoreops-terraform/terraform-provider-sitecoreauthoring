data "sitecoreauthoring_item" "auth_template" {
  path = "/sitecore/templates/System/Webhooks/Authorization/ApiKey "
}

data "sitecoreauthoring_item" "auth_parent" {
  path = "/sitecore/system/Settings/Webhooks/Authorizations"
}

resource "sitecoreauthoring_item" "auth" {
  parent_id   = data.sitecoreauthoring_item.auth_parent.item.item_id
  template_id = data.sitecoreauthoring_item.auth_template.item.item_id
  name        = "Demo"
  language    = "en"
  fields = {
    Key   = "x-secret-header"
    Value = "secret-value"
  }
}

data "sitecoreauthoring_item" "submit_action_template" {
  path = "/sitecore/templates/System/Workflow/Webhooks/Webhook Submit Action"
}
data "sitecoreauthoring_item" "approve_action_parent" {
  path = "/sitecore/system/Workflows/Sample Workflow/Awaiting Approval/Approve"
}
resource "sitecoreauthoring_item" "approve_action_webhook" {
  parent_id   = data.sitecoreauthoring_item.approve_action_parent.item.item_id
  template_id = data.sitecoreauthoring_item.submit_action_template.item.item_id
  name        = "Webhook"
  language    = "en"
  fields = {
    Description   = "My webhook"
    Enabled       = "1"
    Url           = "https://my-webhook-in-terraform"
    Authorization = sitecoreauthoring_item.auth.item_id
  }
}
