
# Get all sites
data "sitecoreauthoring_sites" "all" {}

# Output site information
output "site_names" {
  value = [for site in data.sitecoreauthoring_sites.all.sites : site.name]
}

output "site_details" {
  value = data.sitecoreauthoring_sites.all.sites
}