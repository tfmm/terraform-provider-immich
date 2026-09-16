package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLibraryClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/libraries" {
				w.Write([]byte(`[{"id":"lib-1","name":"Main Library"}]`))
				return
			}
			if r.URL.Path == "/libraries/lib-1" {
				w.Write([]byte(`{"id":"lib-1","name":"Main Library"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"lib-2","name":"New Library"}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"lib-2","name":"Updated Library"}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	libs, err := c.GetLibraries(context.Background())
	if err != nil || len(libs) != 1 {
		t.Fatalf("GetLibraries failed: %v", err)
	}

	lib, err := c.GetLibrary(context.Background(), "lib-1")
	if err != nil || lib.Name != "Main Library" {
		t.Fatalf("GetLibrary failed: %v", err)
	}

	created, err := c.CreateLibrary(context.Background(), CreateLibraryRequest{Name: "New Library"})
	if err != nil || created.ID != "lib-2" {
		t.Fatalf("CreateLibrary failed: %v", err)
	}

	updated, err := c.UpdateLibrary(context.Background(), "lib-2", UpdateLibraryRequest{Name: "Updated Library"})
	if err != nil || updated.Name != "Updated Library" {
		t.Fatalf("UpdateLibrary failed: %v", err)
	}

	err = c.DeleteLibrary(context.Background(), "lib-2")
	if err != nil {
		t.Fatalf("DeleteLibrary failed: %v", err)
	}
}
