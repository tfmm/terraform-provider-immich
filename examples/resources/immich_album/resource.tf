resource "immich_album" "example" {
  name        = "Vacation 2024"
  description = "Photos from our summer vacation"
  order       = "desc"

  users = [
    {
      user_id = "some-user-id"
      role    = "Editor"
    }
  ]
}
