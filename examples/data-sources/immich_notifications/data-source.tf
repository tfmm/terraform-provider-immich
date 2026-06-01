data "immich_notifications" "my_unread" {
  unread_only = true
}

output "unread_notifications" {
  value = data.immich_notifications.my_unread.notifications
}
