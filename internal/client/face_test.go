package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFaceClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`[{"id":"face-1","assetId":"asset-1","personId":"person-1"}]`))
		case "POST":
			w.Write([]byte(`{"id":"face-2","assetId":"asset-1","personId":"person-1"}`))
		case "PUT":
			w.WriteHeader(http.StatusOK)
		case "DELETE":
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	faces, err := c.GetFaces(context.Background(), "asset-1")
	if err != nil || len(faces) != 1 {
		t.Fatalf("GetFaces failed: %v", err)
	}

	created, err := c.CreateFace(context.Background(), CreateFaceRequest{AssetId: "asset-1", PersonId: "person-1"})
	if err != nil || created.ID != "face-2" {
		t.Fatalf("CreateFace failed: %v", err)
	}

	updated, err := c.UpdateFace(context.Background(), "face-2", UpdateFaceRequest{PersonId: "person-2"})
	if err != nil || updated.PersonId != "person-2" {
		t.Fatalf("UpdateFace failed: %v", err)
	}

	err = c.DeleteFace(context.Background(), "face-2")
	if err != nil {
		t.Fatalf("DeleteFace failed: %v", err)
	}
}
