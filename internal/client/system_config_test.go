package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSystemConfigClient(t *testing.T) {
	t.Run("uses /admin/config primary", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/admin/config" {
				if r.Method == "GET" {
					w.Write([]byte(`{"server":{"enabled":true}}`))
					return
				}
				if r.Method == "PUT" {
					w.Write([]byte(`{"server":{"enabled":false}}`))
					return
				}
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := NewClient(server.URL, "token")

		cfg, err := c.GetSystemConfig(context.Background())
		if err != nil || cfg.Server["enabled"] != true {
			t.Fatalf("GetSystemConfig failed: %v", err)
		}

		updated, err := c.UpdateSystemConfig(context.Background(), *cfg)
		if err != nil || updated.Server["enabled"] != false {
			t.Fatalf("UpdateSystemConfig failed: %v", err)
		}
	})

	t.Run("fallbacks to /system-config on 404", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/system-config" {
				if r.Method == "GET" {
					w.Write([]byte(`{"theme":{"customCss":".test{}"}}`))
					return
				}
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		c := NewClient(server.URL, "token")

		cfg, err := c.GetSystemConfig(context.Background())
		if err != nil || cfg.Theme["customCss"] != ".test{}" {
			t.Fatalf("GetSystemConfig fallback failed: %v", err)
		}
	})
}
