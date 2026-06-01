data "immich_activities" "album_activities" {
  album_id = "your-album-uuid"
}

output "all_comments" {
  value = [for a in data.immich_activities.album_activities.activities : a.comment if a.type == "COMMENT"]
}

output "like_count" {
  value = length([for a in data.immich_activities.album_activities.activities : a if a.type == "LIKE"])
}
