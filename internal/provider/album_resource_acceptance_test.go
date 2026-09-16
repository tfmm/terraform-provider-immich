package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tfstate "github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// fakeAlbumServer is a minimal, stateful in-memory stand-in for the parts of
// the Immich API that internal/client/album.go talks to. Acceptance tests
// run real plan/apply/refresh cycles through Terraform core against it, so
// unlike the unit tests it also catches divergence between what the config
// says and what actually ends up in the (fake) backend across multiple
// applies - which is exactly the shape of bug that shipped in the album
// resource's Update() (see REVIEW.md finding #1).
type fakeAlbumServer struct {
	mu     sync.Mutex
	albums map[string]*client.Album
	nextID int

	// call counts, for asserting the right endpoints were actually hit.
	addAssetsCalls    int
	removeAssetsCalls int
	addUsersCalls     int
	removeUserCalls   int
	updateRoleCalls   int
}

func newFakeAlbumServer() (*fakeAlbumServer, *httptest.Server) {
	s := &fakeAlbumServer{albums: map[string]*client.Album{}}
	return s, httptest.NewServer(http.HandlerFunc(s.handle))
}

func (s *fakeAlbumServer) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/albums")

	switch {
	case r.Method == http.MethodPost && path == "":
		s.create(w, r)
	case r.Method == http.MethodGet && !strings.Contains(strings.TrimPrefix(path, "/"), "/"):
		s.get(w, strings.TrimPrefix(path, "/"))
	case r.Method == http.MethodPatch && !strings.Contains(strings.TrimPrefix(path, "/"), "/"):
		s.update(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == http.MethodDelete && !strings.Contains(strings.TrimPrefix(path, "/"), "/"):
		s.delete(w, strings.TrimPrefix(path, "/"))
	case r.Method == http.MethodPut && strings.HasSuffix(path, "/assets"):
		s.addAssetsCalls++
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/assets")
		s.get(w, id)
	case r.Method == http.MethodDelete && strings.HasSuffix(path, "/assets"):
		s.removeAssetsCalls++
		w.Write([]byte(`{}`))
	case r.Method == http.MethodPut && strings.HasSuffix(path, "/users"):
		s.addUsersCalls++
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/users")
		s.addUsers(w, r, id)
	case r.Method == http.MethodPut && strings.Contains(path, "/user/"):
		s.updateRoleCalls++
		parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/user/", 2)
		s.updateUserRole(w, r, parts[0], parts[1])
	case r.Method == http.MethodDelete && strings.Contains(path, "/user/"):
		s.removeUserCalls++
		parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/user/", 2)
		s.removeUser(w, parts[0], parts[1])
	default:
		http.NotFound(w, r)
	}
}

func (s *fakeAlbumServer) create(w http.ResponseWriter, r *http.Request) {
	var reqBody client.CreateAlbumRequest
	_ = json.NewDecoder(r.Body).Decode(&reqBody)

	s.nextID++
	album := &client.Album{
		ID:                    fmt.Sprintf("album-%d", s.nextID),
		AlbumName:             reqBody.AlbumName,
		Description:           reqBody.Description,
		AlbumThumbnailAssetId: nil,
		IsActivityEnabled:     false,
		Order:                 "desc",
	}
	for _, u := range reqBody.AlbumUsers {
		album.AlbumUsers = append(album.AlbumUsers, client.AlbumUser{
			User: &client.User{ID: u.UserId},
			Role: u.Role,
		})
	}
	album.AssetCount = len(reqBody.AssetIds)

	s.albums[album.ID] = album
	json.NewEncoder(w).Encode(album)
}

func (s *fakeAlbumServer) get(w http.ResponseWriter, id string) {
	album, ok := s.albums[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
		return
	}
	json.NewEncoder(w).Encode(album)
}

func (s *fakeAlbumServer) update(w http.ResponseWriter, r *http.Request, id string) {
	album, ok := s.albums[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var reqBody client.UpdateAlbumRequest
	_ = json.NewDecoder(r.Body).Decode(&reqBody)

	if reqBody.AlbumName != "" {
		album.AlbumName = reqBody.AlbumName
	}
	album.Description = reqBody.Description
	album.AlbumThumbnailAssetId = reqBody.AlbumThumbnailAssetId
	if reqBody.IsActivityEnabled != nil {
		album.IsActivityEnabled = *reqBody.IsActivityEnabled
	}
	if reqBody.Order != "" {
		album.Order = reqBody.Order
	}
	json.NewEncoder(w).Encode(album)
}

func (s *fakeAlbumServer) delete(w http.ResponseWriter, id string) {
	delete(s.albums, id)
	w.Write([]byte(`{}`))
}

func (s *fakeAlbumServer) addUsers(w http.ResponseWriter, r *http.Request, id string) {
	album, ok := s.albums[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var reqBody client.AddUsersRequest
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	for _, u := range reqBody.AlbumUsers {
		album.AlbumUsers = append(album.AlbumUsers, client.AlbumUser{
			User: &client.User{ID: u.UserId},
			Role: u.Role,
		})
	}
	json.NewEncoder(w).Encode(album)
}

func (s *fakeAlbumServer) updateUserRole(w http.ResponseWriter, r *http.Request, albumID, userID string) {
	album, ok := s.albums[albumID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var reqBody client.UpdateAlbumUserRequest
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	for i, u := range album.AlbumUsers {
		if u.User != nil && u.User.ID == userID {
			album.AlbumUsers[i].Role = reqBody.Role
		}
	}
	json.NewEncoder(w).Encode(album)
}

func (s *fakeAlbumServer) removeUser(w http.ResponseWriter, albumID, userID string) {
	album, ok := s.albums[albumID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	kept := album.AlbumUsers[:0]
	for _, u := range album.AlbumUsers {
		if u.User == nil || u.User.ID != userID {
			kept = append(kept, u)
		}
	}
	album.AlbumUsers = kept
	w.Write([]byte(`{}`))
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"immich": providerserver.NewProtocol6WithError(New("acctest")()),
	}
}

// TestAccAlbumResource_UsersAndAssetIdsConverge is an acceptance-level
// regression test for the bug fixed alongside REVIEW.md finding #1: editing
// `users` or `asset_ids` on an existing album previously had no effect on
// the backend, so state never converged with config. terraform-plugin-testing
// automatically re-plans after each apply step and fails if that plan is
// non-empty, which would have caught the `users` half of that bug; the
// fake server's call counters directly verify the `asset_ids` half, since
// the album API never returns an asset list to refresh state from.
func TestAccAlbumResource_UsersAndAssetIdsConverge(t *testing.T) {
	store, server := newFakeAlbumServer()
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
resource "immich_album" "test" {
  name        = "Test Album"
  description = "Initial description"
  asset_ids   = ["asset-old"]
  users = [
    { user_id = "user-old", role = "viewer" },
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("immich_album.test", "name", "Test Album"),
					resource.TestCheckResourceAttr("immich_album.test", "users.0.user_id", "user-old"),
					resource.TestCheckResourceAttr("immich_album.test", "users.0.role", "viewer"),
				),
			},
			{
				// Changes name/description (already worked before the fix)
				// AND users/asset_ids (previously silently dropped).
				Config: providerConfig + `
resource "immich_album" "test" {
  name        = "Renamed Album"
  description = "Updated description"
  asset_ids   = ["asset-old", "asset-new"]
  users = [
    { user_id = "user-old", role = "editor" },
    { user_id = "user-new", role = "viewer" },
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("immich_album.test", "name", "Renamed Album"),
					resource.TestCheckResourceAttr("immich_album.test", "users.#", "2"),
					func(_ *tfstate.State) error {
						store.mu.Lock()
						defer store.mu.Unlock()
						if store.addAssetsCalls == 0 {
							return fmt.Errorf("expected AddAssetsToAlbum to have been called")
						}
						if store.addUsersCalls == 0 {
							return fmt.Errorf("expected AddUsersToAlbum to have been called")
						}
						if store.updateRoleCalls == 0 {
							return fmt.Errorf("expected UpdateAlbumUserRole to have been called")
						}
						return nil
					},
				),
				// The framework re-plans after this step and fails the test
				// if that plan is non-empty. Before the fix, `users` would
				// have reverted in state to the pre-update value (since the
				// backend was never actually updated), producing a diff
				// here and failing this exact assertion.
			},
		},
	})
}
