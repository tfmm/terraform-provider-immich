resource "immich_api_key" "example" {
  name        = "Example API Key"
  permissions = ["asset.read", "asset.upload"]
}
