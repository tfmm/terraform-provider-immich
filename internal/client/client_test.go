package client

import (
	"errors"
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

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("expected *APIError, got %T: %v", err, err)
		}
		if apiErr.StatusCode != http.StatusBadRequest {
			t.Errorf("expected StatusCode 400, got %d", apiErr.StatusCode)
		}
		if IsNotFound(err) {
			t.Errorf("expected IsNotFound to be false for a 400")
		}
	})

	t.Run("returns a 404 APIError that IsNotFound recognizes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Not Found"}`))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-api-key")
		req, err := http.NewRequest("GET", server.URL+"/missing", nil)
		if err != nil {
			t.Fatalf("unexpected request creation error: %v", err)
		}

		_, err = c.doRequest(req)
		if err == nil {
			t.Fatalf("expected error for HTTP 404, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected IsNotFound to be true for a 404, got error: %v", err)
		}
	})
}
