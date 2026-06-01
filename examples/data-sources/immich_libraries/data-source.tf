data "immich_libraries" "all" {}

output "library_names" {
  value = data.immich_libraries.all.libraries[*].name
}

output "external_libraries" {
  value = [for l in data.immich_libraries.all.libraries : l if l.type == "EXTERNAL"]
}
