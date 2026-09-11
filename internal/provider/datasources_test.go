package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestAllDataSourcesMetadataAndSchema(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		datasourceFunc func() datasource.DataSource
		expectedType   string
		mainAttr       string
	}{
		{"UsersDataSource", NewUsersDataSource, "immich_users", "users"},
		{"AlbumsDataSource", NewAlbumsDataSource, "immich_albums", "albums"},
		{"LibrariesDataSource", NewLibrariesDataSource, "immich_libraries", "libraries"},
		{"ActivitiesDataSource", NewActivitiesDataSource, "immich_activities", "activities"},
		{"ServerDataSource", NewServerDataSource, "immich_server_info", "id"},
		{"NotificationsDataSource", NewNotificationsDataSource, "immich_notifications", "notifications"},
		{"FacesDataSource", NewFacesDataSource, "immich_faces", "faces"},
		{"AssetsDataSource", NewAssetsDataSource, "immich_assets", "assets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := tt.datasourceFunc()

			metaReq := datasource.MetadataRequest{ProviderTypeName: "immich"}
			metaResp := &datasource.MetadataResponse{}
			ds.Metadata(ctx, metaReq, metaResp)

			if metaResp.TypeName != tt.expectedType {
				t.Errorf("expected datasource type name %q, got %q", tt.expectedType, metaResp.TypeName)
			}

			schemaReq := datasource.SchemaRequest{}
			schemaResp := &datasource.SchemaResponse{}
			ds.Schema(ctx, schemaReq, schemaResp)

			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("unexpected schema diagnostics error: %v", schemaResp.Diagnostics)
			}

			if _, ok := schemaResp.Schema.Attributes[tt.mainAttr]; !ok {
				t.Errorf("expected datasource schema to contain %q attribute", tt.mainAttr)
			}
		})
	}
}
