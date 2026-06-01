resource "immich_activity" "album_comment" {
  album_id = "your-album-uuid"
  type     = "comment"
  comment  = "This is a great album!"
}

resource "immich_activity" "asset_like" {
  album_id = "your-album-uuid"
  asset_id = "your-asset-uuid"
  type     = "like"
}
