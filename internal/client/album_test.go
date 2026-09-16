package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAlbumJSONUnmarshal(t *testing.T) {
	jsonInput := `{"id":"album-1","albumName":"Test Album","description":null,"isActivityEnabled":true}`

	var a Album
	if err := json.Unmarshal([]byte(jsonInput), &a); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if a.Description != nil {
		t.Errorf("expected nil Description for null json, got %v", *a.Description)
	}
}

func TestCreateAlbumRequestJSON(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	req := CreateAlbumRequest{
		AlbumName:   "My Album",
		Description: strPtr("Album description"),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	expected := `{"albumName":"My Album","description":"Album description"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestAlbumClientHTTPMethods(t *testing.T) {
	t.Run("GetAlbums & GetAlbum", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/albums" {
				w.Write([]byte(`[{"id":"alb-1","albumName":"Album 1"}]`))
				return
			}
			if r.URL.Path == "/albums/alb-1" {
				w.Write([]byte(`{"id":"alb-1","albumName":"Album 1"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := NewClient(server.URL, "key")
		albums, err := c.GetAlbums(context.Background())
		if err != nil || len(albums) != 1 {
			t.Fatalf("GetAlbums failed: %v", err)
		}

		album, err := c.GetAlbum(context.Background(), "alb-1")
		if err != nil || album.AlbumName != "Album 1" {
			t.Fatalf("GetAlbum failed: %v", err)
		}
	})

	t.Run("CreateAlbum, UpdateAlbum, DeleteAlbum", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "POST":
				w.Write([]byte(`{"id":"alb-2","albumName":"New Album"}`))
			case "PATCH":
				w.Write([]byte(`{"id":"alb-2","albumName":"Updated Album"}`))
			case "DELETE":
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		c := NewClient(server.URL, "key")
		created, err := c.CreateAlbum(context.Background(), CreateAlbumRequest{AlbumName: "New Album"})
		if err != nil || created.ID != "alb-2" {
			t.Fatalf("CreateAlbum failed: %v", err)
		}

		updated, err := c.UpdateAlbum(context.Background(), "alb-2", UpdateAlbumRequest{AlbumName: "Updated Album"})
		if err != nil || updated.AlbumName != "Updated Album" {
			t.Fatalf("UpdateAlbum failed: %v", err)
		}

		err = c.DeleteAlbum(context.Background(), "alb-2")
		if err != nil {
			t.Fatalf("DeleteAlbum failed: %v", err)
		}
	})
}
