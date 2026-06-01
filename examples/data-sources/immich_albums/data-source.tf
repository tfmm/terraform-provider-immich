data "immich_albums" "all" {}

output "album_names" {
  value = data.immich_albums.all.albums[*].name
}
