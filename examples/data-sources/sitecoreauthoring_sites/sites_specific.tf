# Example: Get specific sites by name
# This example shows how to retrieve specific sites by name

# Get specific sites by name
data "sitecoreauthoring_sites" "specific" {
  site_names = ["reference"]
}

# Output site information
output "specific_sites" {
  value = data.sitecoreauthoring_sites.specific.sites
}