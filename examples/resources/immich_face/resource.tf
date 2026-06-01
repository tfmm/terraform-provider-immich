resource "immich_face" "manual" {
  asset_id        = "your-asset-uuid"
  person_id       = "your-person-uuid"
  bounding_box_x1 = 100.5
  bounding_box_y1 = 100.5
  bounding_box_x2 = 200.5
  bounding_box_y2 = 200.5
  image_height    = 1080
  image_width     = 1920
}
