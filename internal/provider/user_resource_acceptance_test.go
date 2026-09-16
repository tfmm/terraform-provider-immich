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

// fakeUserServer is a minimal in-memory stand-in for the Immich admin users
// API, tracking the last password each request actually sent so tests can
// assert rotation only happens when password_wo_version changes.
type fakeUserServer struct {
	mu           sync.Mutex
	users        map[string]*client.User
	nextID       int
	lastPassword string
	updateCalls  int
}

func newFakeUserServer() (*fakeUserServer, *httptest.Server) {
	s := &fakeUserServer{users: map[string]*client.User{}}
	return s, httptest.NewServer(http.HandlerFunc(s.handle))
}

func (s *fakeUserServer) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/admin/users")

	switch {
	case r.Method == http.MethodPost && path == "":
		var reqBody client.UserAdminCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		s.nextID++
		s.lastPassword = reqBody.Password
		user := &client.User{
			ID:      fmt.Sprintf("user-%d", s.nextID),
			Email:   reqBody.Email,
			Name:    reqBody.Name,
			IsAdmin: reqBody.IsAdmin,
		}
		s.users[user.ID] = user
		json.NewEncoder(w).Encode(user)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/"):
		id := strings.TrimPrefix(path, "/")
		user, ok := s.users[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(user)
	case r.Method == http.MethodPut && strings.HasPrefix(path, "/"):
		id := strings.TrimPrefix(path, "/")
		user, ok := s.users[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var reqBody client.UserAdminUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody.Name != "" {
			user.Name = reqBody.Name
		}
		if reqBody.Password != "" {
			s.updateCalls++
			s.lastPassword = reqBody.Password
		}
		json.NewEncoder(w).Encode(user)
	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/"):
		id := strings.TrimPrefix(path, "/")
		delete(s.users, id)
		w.Write([]byte(`{}`))
	default:
		http.NotFound(w, r)
	}
}

// TestAccUserResource_PasswordWriteOnlyRotation is an acceptance-level test
// for REVIEW.md finding #8: `password_wo` must never be persisted to state,
// and the password should only be re-sent to the API when
// `password_wo_version` actually changes - not on every apply.
func TestAccUserResource_PasswordWriteOnlyRotation(t *testing.T) {
	store, server := newFakeUserServer()
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
resource "immich_user" "test" {
  email                = "test@example.com"
  name                 = "Test User"
  password_wo          = "initial-secret"
  password_wo_version  = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("immich_user.test", "name", "Test User"),
					resource.TestCheckNoResourceAttr("immich_user.test", "password_wo"),
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
				// Same version: name changes, but the password must NOT be
				// re-sent since password_wo_version didn't change.
				Config: providerConfig + `
resource "immich_user" "test" {
  email                = "test@example.com"
  name                 = "Renamed User"
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
				// Version bump: password must be rotated even though the
				// literal password_wo value can't be diffed (write-only).
				Config: providerConfig + `
resource "immich_user" "test" {
  email                = "test@example.com"
  name                 = "Renamed User"
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
