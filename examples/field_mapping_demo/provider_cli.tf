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
  # Authenticate with CLI before running terraform
  # The .sitecore folder must be in terraform folder or a parent folder
  # Initialize Sitecore CLI:
  # > dotnet tool install Sitecore.CLI 
  # > dotnet sitecore init
  # > dotnet sitecore plugin add -n Sitecore.DevEx.Extensibility.XMCloud
  # Plugin documentation: https://doc.sitecore.com/sai/en/developers/sitecoreai/the-cli-cloud-command.html
  # Authenticate by running
  # > dotnet sitecore cloud login
  use_cli      = true
  cli_endpoint = "dev"
}