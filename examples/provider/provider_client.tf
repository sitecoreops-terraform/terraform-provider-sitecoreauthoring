terraform {
  required_providers {
    sitecoreauthoring = {
      source = "sitecoreops-terraform/sitecoreauthoring"
    }
  }
  required_version = ">= 0.0.1"
}

# Configure the Sitecore AI provider
provider "sitecoreauthoring" {
  client_id     = "your autonation client id"
  client_secret = "your autonation client secret"
  host          = "https://xmc-you-authoring-host.sitecorecloud.io"
}
