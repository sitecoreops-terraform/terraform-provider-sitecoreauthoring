# Example: Search for a single site by name
# This example shows how to search for a specific site by name

# Configure the provider
# Search for a single site
data "sitecoreauthoring_sites" "search" {
  search_name = "website"
}

# Output site information
output "search_found_site" {
  value = data.sitecoreauthoring_sites.search.sites
}

# Output specific properties
output "search_site_found" {
  value = length(data.sitecoreauthoring_sites.search.sites) > 0 ? data.sitecoreauthoring_sites.search.sites[0].root_path : "Not found"
}