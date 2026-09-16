package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDoRequestHeadersAndErrors(t *testing.T) {
	t.Run("sets default headers correctly", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey := r.Header.Get("x-api-key"); apiKey != "test-api-key" {
				t.Errorf("expected x-api-key 'test-api-key', got %q", apiKey)
			}
			if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", contentType)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-api-key")
		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		if err != nil {
			t.Fatalf("unexpected request creation error: %v", err)
		}

		body, err := c.doRequest(req)
		if err != nil {
			t.Fatalf("unexpected doRequest error: %v", err)
		}
		if string(body) != `{"status":"ok"}` {
			t.Errorf("expected body '{\"status\":\"ok\"}', got %q", string(body))
		}
	})

	t.Run("returns error on non-2xx status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Bad Request"}`))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-api-key")
		req, err := http.NewRequest("GET", server.URL+"/error", nil)
		if err != nil {
			t.Fatalf("unexpected request creation error: %v", err)
		}

		_, err = c.doRequest(req)
		if err == nil {
			t.Fatalf("expected error for HTTP 400, got nil")
		}
	})
}
