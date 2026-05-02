data "sitecoreauthoring_sites" "byrootpath" {
  root_path = "/sitecore/content/asmblii"
}

# Output site information
output "byrootpath_site_names" {
  value = [for site in data.sitecoreauthoring_sites.byrootpath.sites : site.name]
}
