data "immich_server_info" "current" {}

output "server_version" {
  value = data.immich_server_info.current.version
}
