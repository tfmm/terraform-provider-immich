resource "immich_user" "example" {
  email    = "user@example.com"
  name     = "Example User"
  password = "securepassword123"
  is_admin = false
}
