package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTagsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/tags" {
				w.Write([]byte(`[{"id":"tag-1","name":"Nature"}]`))
				return
			}
			if r.URL.Path == "/tags/tag-1" {
				w.Write([]byte(`{"id":"tag-1","name":"Nature"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"tag-2","name":"Travel"}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"tag-2","name":"Vacation"}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	tags, err := c.GetTags(context.Background())
	if err != nil || len(tags) != 1 {
		t.Fatalf("GetTags failed: %v", err)
	}

	tag, err := c.GetTag(context.Background(), "tag-1")
	if err != nil || tag.Name != "Nature" {
		t.Fatalf("GetTag failed: %v", err)
	}

	created, err := c.CreateTag(context.Background(), CreateTagRequest{Name: "Travel"})
	if err != nil || created.ID != "tag-2" {
		t.Fatalf("CreateTag failed: %v", err)
	}

	color := "#ff0000"
	updated, err := c.UpdateTag(context.Background(), "tag-2", UpdateTagRequest{Color: &color})
	if err != nil || updated.Name != "Vacation" {
		t.Fatalf("UpdateTag failed: %v", err)
	}

	err = c.DeleteTag(context.Background(), "tag-2")
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}
}
