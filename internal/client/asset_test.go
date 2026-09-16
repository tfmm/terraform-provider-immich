package client

import (
	"bytes"
	"context"
	"io"
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

	asset, err := c.GetAsset(context.Background(), "asset-1")
	if err != nil || asset.OriginalFileName != "test.jpg" {
		t.Fatalf("GetAsset failed: %v", err)
	}

	updated, err := c.UpdateAsset(context.Background(), "asset-1", UpdateAssetRequest{Description: "Updated desc"})
	if err != nil || updated.Description != "Updated desc" {
		t.Fatalf("UpdateAsset failed: %v", err)
	}

	searchRes, err := c.SearchAssets(context.Background(), SearchAssetsRequest{OriginalFileName: "test.jpg"})
	if err != nil || searchRes.Assets.Total != 1 {
		t.Fatalf("SearchAssets failed: %v", err)
	}

	err = c.DeleteAssets(context.Background(), []string{"asset-1"})
	if err != nil {
		t.Fatalf("DeleteAssets failed: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "dummy.jpg")
	if err := os.WriteFile(tmpFile, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}

	uploaded, err := c.UploadAsset(context.Background(), tmpFile, time.Now(), time.Now(), false)
	if err != nil || uploaded.ID != "asset-2" {
		t.Fatalf("UploadAsset failed: %v", err)
	}
}

// TestUploadAssetStreamsMultipartBodyCorrectly is a regression test for the
// switch from buffering the whole file in a bytes.Buffer to streaming it
// through an io.Pipe: it verifies the server still receives a well-formed
// multipart request with the correct file content and form fields.
func TestUploadAssetStreamsMultipartBodyCorrectly(t *testing.T) {
	wantContent := bytes.Repeat([]byte("streamed-upload-content"), 10_000)

	var gotFileName string
	var gotContent []byte
	var gotIsFavorite string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/assets" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("failed to parse multipart form: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("assetData")
		if err != nil {
			t.Errorf("missing assetData file part: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		gotFileName = header.Filename
		gotContent, err = io.ReadAll(file)
		if err != nil {
			t.Errorf("failed reading uploaded file: %v", err)
		}
		gotIsFavorite = r.FormValue("isFavorite")

		w.Write([]byte(`{"id":"asset-streamed"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	tmpFile := filepath.Join(t.TempDir(), "large.bin")
	if err := os.WriteFile(tmpFile, wantContent, 0644); err != nil {
		t.Fatalf("failed creating temp file: %v", err)
	}

	uploaded, err := c.UploadAsset(context.Background(), tmpFile, time.Now(), time.Now(), true)
	if err != nil {
		t.Fatalf("UploadAsset failed: %v", err)
	}
	if uploaded.ID != "asset-streamed" {
		t.Errorf("expected uploaded asset ID 'asset-streamed', got %q", uploaded.ID)
	}
	if gotFileName != "large.bin" {
		t.Errorf("expected uploaded filename 'large.bin', got %q", gotFileName)
	}
	if !bytes.Equal(gotContent, wantContent) {
		t.Errorf("uploaded content mismatch: got %d bytes, want %d bytes", len(gotContent), len(wantContent))
	}
	if gotIsFavorite != "true" {
		t.Errorf("expected isFavorite 'true', got %q", gotIsFavorite)
	}
}
