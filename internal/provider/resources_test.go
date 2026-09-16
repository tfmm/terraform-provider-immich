package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestAllResourcesMetadataAndSchema(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		resourceFunc func() resource.Resource
		expectedType string
		idAttr       string
	}{
		{"UserResource", NewUserResource, "immich_user", "id"},
		{"ApiKeyResource", NewApiKeyResource, "immich_api_key", "id"},
		{"SharedLinkResource", NewSharedLinkResource, "immich_shared_link", "id"},
		{"AlbumResource", NewAlbumResource, "immich_album", "id"},
		{"SystemConfigResource", NewSystemConfigResource, "immich_system_config", "id"},
		{"LibraryResource", NewLibraryResource, "immich_library", "id"},
		{"ActivityResource", NewActivityResource, "immich_activity", "id"},
		{"PersonResource", NewPersonResource, "immich_person", "id"},
		{"PartnerResource", NewPartnerResource, "immich_partner", "partner_id"},
		{"MemoryResource", NewMemoryResource, "immich_memory", "id"},
		{"StackResource", NewStackResource, "immich_stack", "id"},
		{"TagResource", NewTagResource, "immich_tag", "id"},
		{"WorkflowResource", NewWorkflowResource, "immich_workflow", "id"},
		{"AdminNotificationResource", NewAdminNotificationResource, "immich_admin_notification", "id"},
		{"FaceResource", NewFaceResource, "immich_face", "id"},
		{"AssetResource", NewAssetResource, "immich_asset", "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.resourceFunc()

			metaReq := resource.MetadataRequest{ProviderTypeName: "immich"}
			metaResp := &resource.MetadataResponse{}
			res.Metadata(ctx, metaReq, metaResp)

			if metaResp.TypeName != tt.expectedType {
				t.Errorf("expected resource type name %q, got %q", tt.expectedType, metaResp.TypeName)
			}

			schemaReq := resource.SchemaRequest{}
			schemaResp := &resource.SchemaResponse{}
			res.Schema(ctx, schemaReq, schemaResp)

			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("unexpected schema diagnostics error: %v", schemaResp.Diagnostics)
			}

			if _, ok := schemaResp.Schema.Attributes[tt.idAttr]; !ok {
				t.Errorf("expected resource schema to contain %q attribute", tt.idAttr)
			}
		})
	}
}

func TestAdminNotificationResourceSchema(t *testing.T) {
	res := NewAdminNotificationResource()
	schemaResp := &resource.SchemaResponse{}
	res.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	attrs := schemaResp.Schema.Attributes
	if _, ok := attrs["user_id"]; !ok {
		t.Errorf("expected admin_notification schema to contain 'user_id' attribute")
	}
	if _, ok := attrs["title"]; !ok {
		t.Errorf("expected admin_notification schema to contain 'title' attribute")
	}
}
