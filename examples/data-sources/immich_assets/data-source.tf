data "immich_assets" "favorites" {
  is_favorite = true
  type        = "IMAGE"
}

output "favorite_asset_ids" {
  value = data.immich_assets.favorites.assets[*].id
}
