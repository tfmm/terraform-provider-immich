package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tfstate "github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// fakeSharedLinkServer is a minimal in-memory stand-in for the Immich shared
// links API, tracking the last password sent so tests can assert rotation
// only happens when password_wo_version changes.
type fakeSharedLinkServer struct {
	mu           sync.Mutex
	links        map[string]*client.SharedLink
	nextID       int
	lastPassword string
	updateCalls  int
}

func newFakeSharedLinkServer() (*fakeSharedLinkServer, *httptest.Server) {
	s := &fakeSharedLinkServer{links: map[string]*client.SharedLink{}}
	return s, httptest.NewServer(http.HandlerFunc(s.handle))
}

func (s *fakeSharedLinkServer) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/shared-links")

	switch {
	case r.Method == http.MethodPost && path == "":
		var reqBody client.SharedLinkCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		s.nextID++
		if reqBody.Password != nil {
			s.lastPassword = *reqBody.Password
		}
		link := &client.SharedLink{
			ID:            fmt.Sprintf("link-%d", s.nextID),
			Type:          reqBody.Type,
			Description:   reqBody.Description,
			Key:           fmt.Sprintf("key-%d", s.nextID),
			AllowUpload:   reqBody.AllowUpload != nil && *reqBody.AllowUpload,
			AllowDownload: reqBody.AllowDownload == nil || *reqBody.AllowDownload,
			ShowMetadata:  reqBody.ShowMetadata == nil || *reqBody.ShowMetadata,
		}
		s.links[link.ID] = link
		json.NewEncoder(w).Encode(link)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/"):
		id := strings.TrimPrefix(path, "/")
		link, ok := s.links[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(link)
	case r.Method == http.MethodPatch && strings.HasPrefix(path, "/"):
		id := strings.TrimPrefix(path, "/")
		link, ok := s.links[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var reqBody client.SharedLinkUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody.Password != nil {
			s.updateCalls++
			s.lastPassword = *reqBody.Password
		}
		if reqBody.Description != nil {
			link.Description = reqBody.Description
		}
		if reqBody.Slug != nil {
			link.Slug = reqBody.Slug
		}
		if reqBody.ExpiresAt != nil {
			link.ExpiresAt = reqBody.ExpiresAt
		}
		if reqBody.AllowUpload != nil {
			link.AllowUpload = *reqBody.AllowUpload
		}
		if reqBody.AllowDownload != nil {
			link.AllowDownload = *reqBody.AllowDownload
		}
		if reqBody.ShowMetadata != nil {
			link.ShowMetadata = *reqBody.ShowMetadata
		}
		json.NewEncoder(w).Encode(link)
	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/"):
		id := strings.TrimPrefix(path, "/")
		delete(s.links, id)
		w.Write([]byte(`{}`))
	default:
		http.NotFound(w, r)
	}
}

// TestAccSharedLinkResource_PasswordWriteOnlyRotation mirrors
// TestAccUserResource_PasswordWriteOnlyRotation for immich_shared_link's
// password_wo/password_wo_version pair (REVIEW.md finding #8).
func TestAccSharedLinkResource_PasswordWriteOnlyRotation(t *testing.T) {
	store, server := newFakeSharedLinkServer()
	defer server.Close()

	providerConfig := fmt.Sprintf(`
provider "immich" {
  endpoint = %q
  api_key  = "test-key"
}
`, server.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "immich_shared_link" "test" {
  type                 = "INDIVIDUAL"
  asset_ids            = ["asset-1"]
  password_wo          = "initial-secret"
  password_wo_version  = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("immich_shared_link.test", "password_wo"),
					func(_ *tfstate.State) error {
						store.mu.Lock()
						defer store.mu.Unlock()
						if store.lastPassword != "initial-secret" {
							return fmt.Errorf("expected create to send password 'initial-secret', got %q", store.lastPassword)
						}
						return nil
					},
				),
			},
			{
				// Same version: description changes, password must not be
				// re-sent.
				Config: providerConfig + `
resource "immich_shared_link" "test" {
  type                 = "INDIVIDUAL"
  asset_ids            = ["asset-1"]
  description          = "now with a description"
  password_wo          = "initial-secret"
  password_wo_version  = 1
}
`,
				Check: func(_ *tfstate.State) error {
					store.mu.Lock()
					defer store.mu.Unlock()
					if store.updateCalls != 0 {
						return fmt.Errorf("expected no password update when password_wo_version is unchanged, got %d update(s)", store.updateCalls)
					}
					return nil
				},
			},
			{
				// Version bump: password must rotate.
				Config: providerConfig + `
resource "immich_shared_link" "test" {
  type                 = "INDIVIDUAL"
  asset_ids            = ["asset-1"]
  description          = "now with a description"
  password_wo          = "rotated-secret"
  password_wo_version  = 2
}
`,
				Check: func(_ *tfstate.State) error {
					store.mu.Lock()
					defer store.mu.Unlock()
					if store.updateCalls != 1 {
						return fmt.Errorf("expected exactly 1 password update after version bump, got %d", store.updateCalls)
					}
					if store.lastPassword != "rotated-secret" {
						return fmt.Errorf("expected rotated password 'rotated-secret', got %q", store.lastPassword)
					}
					return nil
				},
			},
		},
	})
}
