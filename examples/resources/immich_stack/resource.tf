resource "immich_stack" "example" {
  asset_ids = [
    "asset-uuid-1",
    "asset-uuid-2"
  ]
  primary_asset_id = "asset-uuid-1"
}
