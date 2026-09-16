package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStacksClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/stacks" {
				w.Write([]byte(`[{"id":"st-1","primaryAssetId":"asset-1"}]`))
				return
			}
			if r.URL.Path == "/stacks/st-1" {
				w.Write([]byte(`{"id":"st-1","primaryAssetId":"asset-1"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"st-2","primaryAssetId":"asset-2"}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"st-2","primaryAssetId":"asset-3"}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	stacks, err := c.GetStacks(context.Background())
	if err != nil || len(stacks) != 1 {
		t.Fatalf("GetStacks failed: %v", err)
	}

	stack, err := c.GetStack(context.Background(), "st-1")
	if err != nil || stack.PrimaryAssetId != "asset-1" {
		t.Fatalf("GetStack failed: %v", err)
	}

	created, err := c.CreateStack(context.Background(), CreateStackRequest{AssetIds: []string{"asset-2"}})
	if err != nil || created.ID != "st-2" {
		t.Fatalf("CreateStack failed: %v", err)
	}

	updated, err := c.UpdateStack(context.Background(), "st-2", UpdateStackRequest{PrimaryAssetId: "asset-3"})
	if err != nil || updated.PrimaryAssetId != "asset-3" {
		t.Fatalf("UpdateStack failed: %v", err)
	}

	err = c.DeleteStack(context.Background(), "st-2")
	if err != nil {
		t.Fatalf("DeleteStack failed: %v", err)
	}
}
