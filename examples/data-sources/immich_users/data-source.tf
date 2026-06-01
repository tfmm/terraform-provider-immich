data "immich_users" "all" {}

output "user_emails" {
  value = data.immich_users.all.users[*].email
}
