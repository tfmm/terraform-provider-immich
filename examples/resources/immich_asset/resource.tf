# Upload an asset from a local file
resource "immich_asset" "upload_example" {
  filename    = "path/to/your/photo.jpg"
  description = "Uploaded via Terraform"
  is_favorite = true
}

# Manage metadata of an existing asset
resource "immich_asset" "metadata_example" {
  id          = "existing-asset-uuid"
  is_favorite = true
  is_archived = false
  description = "Updated description"
}
