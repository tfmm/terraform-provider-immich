resource "immich_library" "external" {
  name = "My External Photos"
  type = "EXTERNAL"
  import_paths = [
    "/mnt/photos/vacation_2023",
    "/mnt/photos/family"
  ]
  exclusion_patterns = [
    "**/tmp/**",
    "**/.DS_Store"
  ]
  is_visible = true
}

resource "immich_library" "upload" {
  name = "Personal Uploads"
  type = "UPLOAD"
}
