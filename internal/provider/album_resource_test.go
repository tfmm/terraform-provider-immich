package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

func TestUpdateAlbumResourceModel(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	t.Run("null description in album response maps to StringNull", func(t *testing.T) {
		var model albumResourceModel
		album := &client.Album{
			ID:          "album-123",
			AlbumName:   "Camera Roll",
			Description: nil,
		}

		updateAlbumResourceModel(&model, album)

		if model.ID.ValueString() != "album-123" {
			t.Errorf("expected ID 'album-123', got %q", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Camera Roll" {
			t.Errorf("expected Name 'Camera Roll', got %q", model.Name.ValueString())
		}
		if !model.Description.IsNull() {
			t.Errorf("expected Description to be null, got %v", model.Description)
		}
	})

	t.Run("empty string description in album response maps to StringNull", func(t *testing.T) {
		var model albumResourceModel
		album := &client.Album{
			ID:          "album-123",
			AlbumName:   "Camera Roll",
			Description: strPtr(""),
		}

		updateAlbumResourceModel(&model, album)

		if !model.Description.IsNull() {
			t.Errorf("expected Description to be null for empty string, got %v", model.Description)
		}
	})

	t.Run("non-empty description in album response maps to StringValue", func(t *testing.T) {
		var model albumResourceModel
		album := &client.Album{
			ID:          "album-123",
			AlbumName:   "Vacation",
			Description: strPtr("Summer 2026"),
		}

		updateAlbumResourceModel(&model, album)

		if model.Description.IsNull() || model.Description.IsUnknown() {
			t.Fatalf("expected known Description value")
		}
		if model.Description.ValueString() != "Summer 2026" {
			t.Errorf("expected Description 'Summer 2026', got %q", model.Description.ValueString())
		}
	})

	t.Run("owner role in albumUsers is filtered out to prevent drift", func(t *testing.T) {
		var model albumResourceModel
		album := &client.Album{
			ID:        "album-123",
			AlbumName: "Family Photos",
			AlbumUsers: []client.AlbumUser{
				{
					User: &client.User{ID: "owner-id-123"},
					Role: "owner",
				},
				{
					User: &client.User{ID: "editor-id-456"},
					Role: "editor",
				},
			},
		}

		updateAlbumResourceModel(&model, album)

		if len(model.Users) != 1 {
			t.Fatalf("expected 1 shared user (owner filtered), got %d", len(model.Users))
		}
		if model.Users[0].UserId.ValueString() != "editor-id-456" {
			t.Errorf("expected shared user ID 'editor-id-456', got %q", model.Users[0].UserId.ValueString())
		}
		if model.Users[0].Role.ValueString() != "editor" {
			t.Errorf("expected shared user role 'editor', got %q", model.Users[0].Role.ValueString())
		}
	})
}

func TestDiffAlbumAssetIds(t *testing.T) {
	plan := []types.String{types.StringValue("asset-old"), types.StringValue("asset-new")}
	state := []types.String{types.StringValue("asset-old"), types.StringValue("asset-remove")}

	toAdd, toRemove := diffAlbumAssetIds(plan, state)

	if len(toAdd) != 1 || toAdd[0] != "asset-new" {
		t.Errorf("expected toAdd [asset-new], got %v", toAdd)
	}
	if len(toRemove) != 1 || toRemove[0] != "asset-remove" {
		t.Errorf("expected toRemove [asset-remove], got %v", toRemove)
	}
}

func TestDiffAlbumUsers(t *testing.T) {
	plan := []albumUserModel{
		{UserId: types.StringValue("user-old"), Role: types.StringValue("editor")},
		{UserId: types.StringValue("user-new"), Role: types.StringValue("viewer")},
	}
	state := []albumUserModel{
		{UserId: types.StringValue("user-old"), Role: types.StringValue("viewer")},
		{UserId: types.StringValue("user-remove"), Role: types.StringValue("viewer")},
	}

	toAdd, toRemove, toUpdateRole := diffAlbumUsers(plan, state)

	if len(toAdd) != 1 || toAdd[0].UserId.ValueString() != "user-new" {
		t.Errorf("expected toAdd [user-new], got %v", toAdd)
	}
	if len(toRemove) != 1 || toRemove[0] != "user-remove" {
		t.Errorf("expected toRemove [user-remove], got %v", toRemove)
	}
	if len(toUpdateRole) != 1 || toUpdateRole[0].UserId.ValueString() != "user-old" || toUpdateRole[0].Role.ValueString() != "editor" {
		t.Errorf("expected toUpdateRole [user-old:editor], got %v", toUpdateRole)
	}
}

// TestAlbumResourceUpdateAppliesMembershipChanges is a regression test for a
// bug where changing `users` or `asset_ids` on an existing album silently had
// no effect: Update() only sent name/description/thumbnail/order/activity to
// the API and never called the add/remove endpoints for members or assets.
func TestAlbumResourceUpdateAppliesMembershipChanges(t *testing.T) {
	var sawAddAssets, sawRemoveAssets, sawAddUsers, sawUpdateRole, sawRemoveUser bool
	var addedAssetIDs, removedAssetIDs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/albums/album-1/assets":
			sawAddAssets = true
			var body client.BulkIdsRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			addedAssetIDs = body.Ids
			w.Write([]byte(`{}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/albums/album-1/assets":
			sawRemoveAssets = true
			var body client.BulkIdsRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			removedAssetIDs = body.Ids
			w.Write([]byte(`{}`))
		case r.Method == http.MethodPut && r.URL.Path == "/albums/album-1/users":
			sawAddUsers = true
			w.Write([]byte(`{"id":"album-1","albumName":"Test"}`))
		case r.Method == http.MethodPut && r.URL.Path == "/albums/album-1/user/user-old":
			sawUpdateRole = true
			w.Write([]byte(`{"id":"album-1","albumName":"Test"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/albums/album-1/user/user-remove":
			sawRemoveUser = true
			w.Write([]byte(`{}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/albums/album-1":
			w.Write([]byte(`{"id":"album-1","albumName":"Test","albumUsers":[{"user":{"id":"user-old"},"role":"editor"},{"user":{"id":"user-new"},"role":"viewer"}]}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	r := &albumResource{client: client.NewClient(server.URL, "test-key")}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	sch := schemaResp.Schema

	planModel := albumResourceModel{
		ID:                    types.StringValue("album-1"),
		Name:                  types.StringValue("Test"),
		Description:           types.StringNull(),
		AlbumThumbnailAssetId: types.StringNull(),
		IsActivityEnabled:     types.BoolValue(false),
		Order:                 types.StringNull(),
		AssetIds:              []types.String{types.StringValue("asset-old"), types.StringValue("asset-new")},
		Users: []albumUserModel{
			{UserId: types.StringValue("user-old"), Role: types.StringValue("editor")},
			{UserId: types.StringValue("user-new"), Role: types.StringValue("viewer")},
		},
	}
	stateModel := albumResourceModel{
		ID:                    types.StringValue("album-1"),
		Name:                  types.StringValue("Test"),
		Description:           types.StringNull(),
		AlbumThumbnailAssetId: types.StringNull(),
		IsActivityEnabled:     types.BoolValue(false),
		Order:                 types.StringNull(),
		AssetIds:              []types.String{types.StringValue("asset-old"), types.StringValue("asset-remove")},
		Users: []albumUserModel{
			{UserId: types.StringValue("user-old"), Role: types.StringValue("viewer")},
			{UserId: types.StringValue("user-remove"), Role: types.StringValue("viewer")},
		},
	}

	ctx := context.Background()

	var plan tfsdk.Plan
	plan.Schema = sch
	if diags := plan.Set(ctx, &planModel); diags.HasError() {
		t.Fatalf("unexpected diagnostics setting plan: %v", diags)
	}

	var state tfsdk.State
	state.Schema = sch
	if diags := state.Set(ctx, &stateModel); diags.HasError() {
		t.Fatalf("unexpected diagnostics setting state: %v", diags)
	}

	req := resource.UpdateRequest{Plan: plan, State: state}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: sch}}

	r.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	if !sawAddAssets {
		t.Error("expected AddAssetsToAlbum to be called")
	}
	if !sawRemoveAssets {
		t.Error("expected RemoveAssetsFromAlbum to be called")
	}
	if !sawAddUsers {
		t.Error("expected AddUsersToAlbum to be called")
	}
	if !sawUpdateRole {
		t.Error("expected UpdateAlbumUserRole to be called")
	}
	if !sawRemoveUser {
		t.Error("expected RemoveUserFromAlbum to be called")
	}
	if len(addedAssetIDs) != 1 || addedAssetIDs[0] != "asset-new" {
		t.Errorf("expected added asset IDs [asset-new], got %v", addedAssetIDs)
	}
	if len(removedAssetIDs) != 1 || removedAssetIDs[0] != "asset-remove" {
		t.Errorf("expected removed asset IDs [asset-remove], got %v", removedAssetIDs)
	}
}
