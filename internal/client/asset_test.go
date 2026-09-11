package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAssetClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/assets/asset-1" {
				w.Write([]byte(`{"id":"asset-1","originalFileName":"test.jpg"}`))
				return
			}
		case "PUT":
			if r.URL.Path == "/assets/asset-1" {
				w.Write([]byte(`{"id":"asset-1","description":"Updated desc"}`))
				return
			}
		case "DELETE":
			if r.URL.Path == "/assets" {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "POST":
			if r.URL.Path == "/search/metadata" {
				w.Write([]byte(`{"assets":{"total":1,"count":1,"items":[{"id":"asset-1"}]}}`))
				return
			}
			if r.URL.Path == "/assets" {
				w.Write([]byte(`{"id":"asset-2","originalFileName":"dummy.jpg"}`))
				return
			}
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	asset, err := c.GetAsset("asset-1")
	if err != nil || asset.OriginalFileName != "test.jpg" {
		t.Fatalf("GetAsset failed: %v", err)
	}

	updated, err := c.UpdateAsset("asset-1", UpdateAssetRequest{Description: "Updated desc"})
	if err != nil || updated.Description != "Updated desc" {
		t.Fatalf("UpdateAsset failed: %v", err)
	}

	searchRes, err := c.SearchAssets(SearchAssetsRequest{OriginalFileName: "test.jpg"})
	if err != nil || searchRes.Assets.Total != 1 {
		t.Fatalf("SearchAssets failed: %v", err)
	}

	err = c.DeleteAssets([]string{"asset-1"})
	if err != nil {
		t.Fatalf("DeleteAssets failed: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "dummy.jpg")
	if err := os.WriteFile(tmpFile, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}

	uploaded, err := c.UploadAsset(tmpFile, time.Now(), time.Now(), false)
	if err != nil || uploaded.ID != "asset-2" {
		t.Fatalf("UploadAsset failed: %v", err)
	}
}
