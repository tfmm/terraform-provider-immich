package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPartnersClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`[{"id":"part-1","name":"Partner 1"}]`))
		case "POST":
			var req CreatePartnerRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.SharedWithId != "user-123" {
				t.Errorf("expected sharedWithId 'user-123', got %q", req.SharedWithId)
			}
			w.Write([]byte(`{"id":"part-2","name":"Partner 2"}`))
		case "PUT":
			w.Write([]byte(`{"id":"part-2","inTimeline":true}`))
		case "DELETE":
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	partners, err := c.GetPartners()
	if err != nil || len(partners) != 1 {
		t.Fatalf("GetPartners failed: %v", err)
	}

	created, err := c.CreatePartner("user-123")
	if err != nil || created.ID != "part-2" {
		t.Fatalf("CreatePartner failed: %v", err)
	}

	updated, err := c.UpdatePartner("part-2", UpdatePartnerRequest{InTimeline: true})
	if err != nil || !updated.InTimeline {
		t.Fatalf("UpdatePartner failed: %v", err)
	}

	err = c.DeletePartner("part-2")
	if err != nil {
		t.Fatalf("DeletePartner failed: %v", err)
	}
}
