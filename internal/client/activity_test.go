package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActivityClient(t *testing.T) {
	t.Run("GetActivities", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || r.URL.Path != "/activities" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			if r.URL.Query().Get("albumId") != "album-1" {
				t.Errorf("expected albumId query param 'album-1', got %q", r.URL.Query().Get("albumId"))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":"act-1","type":"COMMENT","albumId":"album-1","comment":"Great picture!"}]`))
		}))
		defer server.Close()

		c := NewClient(server.URL, "key")
		activities, err := c.GetActivities("album-1", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(activities) != 1 || activities[0].ID != "act-1" {
			t.Errorf("expected 1 activity with ID 'act-1', got %v", activities)
		}
	})

	t.Run("CreateActivity", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || r.URL.Path != "/activities" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			var req CreateActivityRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Comment != "Nice!" {
				t.Errorf("expected comment 'Nice!', got %q", req.Comment)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"act-2","type":"COMMENT","albumId":"album-1","comment":"Nice!"}`))
		}))
		defer server.Close()

		c := NewClient(server.URL, "key")
		act, err := c.CreateActivity(CreateActivityRequest{
			Type:    "COMMENT",
			AlbumId: "album-1",
			Comment: "Nice!",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if act.ID != "act-2" {
			t.Errorf("expected activity ID 'act-2', got %q", act.ID)
		}
	})

	t.Run("DeleteActivity", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" || r.URL.Path != "/activities/act-1" {
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL, "key")
		err := c.DeleteActivity("act-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
