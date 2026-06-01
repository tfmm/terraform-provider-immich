resource "immich_workflow" "example" {
  name    = "Auto-Tag Video"
  enabled = true
  
  triggers = jsonencode([
    {
      "type": "asset.upload",
      "options": {}
    }
  ])
  
  filters = jsonencode([
    {
      "type": "asset.type",
      "options": {
        "type": "video"
      }
    }
  ])
  
  actions = jsonencode([
    {
      "type": "asset.tag",
      "options": {
        "tagId": "your-tag-uuid"
      }
    }
  ])
}
