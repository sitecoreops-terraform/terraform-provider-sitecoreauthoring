# Terraform Provider for Sitecore Authoring API

This Terraform provider allows you to interact with the Sitecore Authoring API, which provides GraphQL-based access to Sitecore content management functionality.

## Features

- **Sites Data Source**: Retrieve information about Sitecore sites
- **GraphQL Support**: Full support for Sitecore's GraphQL Authoring API
- **Authentication**: Supports both client credentials and Sitecore CLI authentication

## Installation

### Using the Terraform Registry

```hcl
terraform {
  required_providers {
    sitecoreauthoring = {
      source = "registry.terraform.io/sitecoreops-terraform/sitecoreauthoring"
      version = ">= 0.1.0"
    }
  }
}
```
