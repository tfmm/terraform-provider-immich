resource "immich_admin_notification" "announcement" {
  type        = "SYSTEM"
  level       = "INFO"
  title       = "Maintenance Scheduled"
  description = "Immich will be down for maintenance on Sunday at 2 AM UTC."
}
