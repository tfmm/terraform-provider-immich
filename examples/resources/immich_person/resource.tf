resource "immich_person" "example" {
  id         = "your-person-uuid"
  name       = "John Doe"
  birth_date = "1990-01-01"
  is_hidden  = false
  is_favorite = true
}
