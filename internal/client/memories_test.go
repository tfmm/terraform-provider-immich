package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMemoriesClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/memories" {
				w.Write([]byte(`[{"id":"mem-1","memoryAt":"2026-01-01T00:00:00Z"}]`))
				return
			}
			if r.URL.Path == "/memories/mem-1" {
				w.Write([]byte(`{"id":"mem-1","memoryAt":"2026-01-01T00:00:00Z"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"mem-2","memoryAt":"2026-01-02T00:00:00Z"}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"mem-2","isSaved":true}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	memories, err := c.GetMemories(context.Background())
	if err != nil || len(memories) != 1 {
		t.Fatalf("GetMemories failed: %v", err)
	}

	mem, err := c.GetMemory(context.Background(), "mem-1")
	if err != nil || mem.MemoryAt != "2026-01-01T00:00:00Z" {
		t.Fatalf("GetMemory failed: %v", err)
	}

	created, err := c.CreateMemory(context.Background(), CreateMemoryRequest{MemoryAt: "2026-01-02T00:00:00Z"})
	if err != nil || created.ID != "mem-2" {
		t.Fatalf("CreateMemory failed: %v", err)
	}

	isSaved := true
	updated, err := c.UpdateMemory(context.Background(), "mem-2", UpdateMemoryRequest{IsSaved: &isSaved})
	if err != nil || !updated.IsSaved {
		t.Fatalf("UpdateMemory failed: %v", err)
	}

	err = c.DeleteMemory(context.Background(), "mem-2")
	if err != nil {
		t.Fatalf("DeleteMemory failed: %v", err)
	}
}
