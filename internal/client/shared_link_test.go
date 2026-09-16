package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSharedLinkClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/shared-links" {
				w.Write([]byte(`[{"id":"sl-1","type":"ALBUM"}]`))
				return
			}
			if r.URL.Path == "/shared-links/sl-1" {
				w.Write([]byte(`{"id":"sl-1","type":"ALBUM"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"sl-2","type":"INDIVIDUAL"}`))
			return
		case "PATCH":
			w.Write([]byte(`{"id":"sl-2","allowUpload":true}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	links, err := c.GetSharedLinks(context.Background())
	if err != nil || len(links) != 1 {
		t.Fatalf("GetSharedLinks failed: %v", err)
	}

	link, err := c.GetSharedLink(context.Background(), "sl-1")
	if err != nil || link.Type != "ALBUM" {
		t.Fatalf("GetSharedLink failed: %v", err)
	}

	created, err := c.CreateSharedLink(context.Background(), SharedLinkCreateRequest{Type: "INDIVIDUAL"})
	if err != nil || created.ID != "sl-2" {
		t.Fatalf("CreateSharedLink failed: %v", err)
	}

	allowUpload := true
	updated, err := c.UpdateSharedLink(context.Background(), "sl-2", SharedLinkUpdateRequest{AllowUpload: &allowUpload})
	if err != nil || !updated.AllowUpload {
		t.Fatalf("UpdateSharedLink failed: %v", err)
	}

	err = c.DeleteSharedLink(context.Background(), "sl-2")
	if err != nil {
		t.Fatalf("DeleteSharedLink failed: %v", err)
	}
}
