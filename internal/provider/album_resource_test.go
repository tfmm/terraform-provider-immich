package provider

import (
	"testing"

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
}
