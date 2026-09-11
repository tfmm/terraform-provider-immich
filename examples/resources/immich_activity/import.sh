# Import with album_id and activity_id
# Using Terraform CLI
terraform import immich_activity.example album-uuid/activity-uuid

# Using OpenTofu CLI
tofu import immich_activity.example album-uuid/activity-uuid

# Or import with album_id, asset_id and activity_id
# Using Terraform CLI
terraform import immich_activity.example album-uuid/asset-uuid/activity-uuid

# Using OpenTofu CLI
tofu import immich_activity.example album-uuid/asset-uuid/activity-uuid
