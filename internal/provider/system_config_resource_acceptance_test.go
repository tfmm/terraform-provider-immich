package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tfstate "github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/tfmm/terraform-provider-immich/internal/client"
)

// fakeSystemConfigServer is a minimal in-memory stand-in for Immich's
// singleton /admin/config endpoint. Unlike album/user/shared_link, create
// and update both PUT the same document, so password-send counts are
// cumulative rather than reset between steps.
type fakeSystemConfigServer struct {
	mu                sync.Mutex
	config            client.SystemConfig
	lastPassword      string
	passwordSendCount int
}

func newFakeSystemConfigServer() (*fakeSystemConfigServer, *httptest.Server) {
	s := &fakeSystemConfigServer{}
	return s, httptest.NewServer(http.HandlerFunc(s.handle))
}

func (s *fakeSystemConfigServer) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path != "/admin/config" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(s.config)
	case http.MethodPut:
		var incoming client.SystemConfig
		_ = json.NewDecoder(r.Body).Decode(&incoming)

		if notif, ok := incoming.Notifications["smtp"].(map[string]interface{}); ok {
			if transport, ok := notif["transport"].(map[string]interface{}); ok {
				if pw, ok := transport["password"].(string); ok && pw != "" {
					s.passwordSendCount++
					s.lastPassword = pw
				}
			}
		}

		s.config = incoming
		json.NewEncoder(w).Encode(s.config)
	default:
		http.NotFound(w, r)
	}
}

// TestAccSystemConfigResource_SMTPPasswordWriteOnlyRotation mirrors the
// user/shared_link write-only rotation tests for
// notifications.smtp.password_wo (REVIEW.md finding #8), which lives nested
// two levels deep and is spliced into the generic map[string]interface{}
// payload rather than being a top-level model field.
func TestAccSystemConfigResource_SMTPPasswordWriteOnlyRotation(t *testing.T) {
	store, server := newFakeSystemConfigServer()
	defer server.Close()

	providerConfig := `
provider "immich" {
  endpoint = "` + server.URL + `"
  api_key  = "test-key"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "immich_system_config" "test" {
  notifications = {
    smtp = {
      enabled              = true
      host                 = "smtp.example.com"
      password_wo          = "initial-secret"
      password_wo_version  = 1
    }
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("immich_system_config.test", "notifications.smtp.password_wo"),
					func(_ *tfstate.State) error {
						store.mu.Lock()
						defer store.mu.Unlock()
						if store.passwordSendCount != 1 {
							return fmt.Errorf("expected 1 password send after create, got %d", store.passwordSendCount)
						}
						if store.lastPassword != "initial-secret" {
							return fmt.Errorf("expected create to send password 'initial-secret', got %q", store.lastPassword)
						}
						return nil
					},
				),
			},
			{
				// Same version: host changes, password must not be re-sent.
				Config: providerConfig + `
resource "immich_system_config" "test" {
  notifications = {
    smtp = {
      enabled              = true
      host                 = "smtp2.example.com"
      password_wo          = "initial-secret"
      password_wo_version  = 1
    }
  }
}
`,
				Check: func(_ *tfstate.State) error {
					store.mu.Lock()
					defer store.mu.Unlock()
					if store.passwordSendCount != 1 {
						return fmt.Errorf("expected password send count to stay at 1 when password_wo_version is unchanged, got %d", store.passwordSendCount)
					}
					return nil
				},
			},
			{
				// Version bump: password must rotate.
				Config: providerConfig + `
resource "immich_system_config" "test" {
  notifications = {
    smtp = {
      enabled              = true
      host                 = "smtp2.example.com"
      password_wo          = "rotated-secret"
      password_wo_version  = 2
    }
  }
}
`,
				Check: func(_ *tfstate.State) error {
					store.mu.Lock()
					defer store.mu.Unlock()
					if store.passwordSendCount != 2 {
						return fmt.Errorf("expected password send count 2 after version bump, got %d", store.passwordSendCount)
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
