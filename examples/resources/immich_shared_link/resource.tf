resource "immich_shared_link" "example" {
  type           = "ALBUM"
  album_id       = "some-album-id"
  description    = "My shared album"
  allow_download = true
}
