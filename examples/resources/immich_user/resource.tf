resource "immich_user" "example" {
  email    = "user@example.com"
  name     = "Example User"
  is_admin = false

  # Recommended: write-only, never persisted to plan or state. Bump
  # password_wo_version to rotate the password on a later apply.
  # Requires Terraform 1.11+.
  password_wo         = "securepassword123"
  password_wo_version = 1
}
