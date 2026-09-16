package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiKeyClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/api-keys" {
				w.Write([]byte(`[{"id":"key-1","name":"Test Key"}]`))
				return
			}
			if r.URL.Path == "/api-keys/key-1" {
				w.Write([]byte(`{"id":"key-1","name":"Test Key"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"secret":"secret-value","apiKey":{"id":"key-2","name":"New Key"}}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"key-2","name":"Updated Key"}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	keys, err := c.GetApiKeys()
	if err != nil || len(keys) != 1 {
		t.Fatalf("GetApiKeys failed: %v", err)
	}

	key, err := c.GetApiKey("key-1")
	if err != nil || key.Name != "Test Key" {
		t.Fatalf("GetApiKey failed: %v", err)
	}

	created, err := c.CreateApiKey(ApiKeyCreateRequest{Name: "New Key"})
	if err != nil || created.Secret != "secret-value" || created.ApiKey.ID != "key-2" {
		t.Fatalf("CreateApiKey failed: %v", err)
	}

	updated, err := c.UpdateApiKey("key-2", ApiKeyUpdateRequest{Name: "Updated Key"})
	if err != nil || updated.Name != "Updated Key" {
		t.Fatalf("UpdateApiKey failed: %v", err)
	}

	err = c.DeleteApiKey("key-2")
	if err != nil {
		t.Fatalf("DeleteApiKey failed: %v", err)
	}
}
