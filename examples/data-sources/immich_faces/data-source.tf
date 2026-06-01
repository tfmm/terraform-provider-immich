data "immich_faces" "current" {
  asset_id = "your-asset-uuid"
}

output "detected_faces" {
  value = data.immich_faces.current.faces
}
