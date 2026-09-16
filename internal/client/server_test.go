package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/server/about":
			w.Write([]byte(`{"version":"v1.100.0","build":"main"}`))
		case "/public/config":
			w.Write([]byte(`{"oauth":{"enabled":true},"passwordLogin":{"enabled":false}}`))
		case "/server/statistics":
			w.Write([]byte(`{"photos":100,"videos":10,"usage":10000,"users":2}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	about, err := c.GetServerAbout(context.Background())
	if err != nil || about.Version != "v1.100.0" {
		t.Fatalf("GetServerAbout failed: %v", err)
	}

	features, err := c.GetServerFeatures(context.Background())
	if err != nil || !features.Oauth || features.PasswordLogin {
		t.Fatalf("GetServerFeatures failed: %v", err)
	}

	stats, err := c.GetServerStatistics(context.Background())
	if err != nil || stats.Photos != 100 {
		t.Fatalf("GetServerStatistics failed: %v", err)
	}
}
