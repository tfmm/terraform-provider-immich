package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkflowsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/workflows" {
				w.Write([]byte(`[{"id":"wf-1","name":"Workflow 1","enabled":true}]`))
				return
			}
			if r.URL.Path == "/workflows/wf-1" {
				w.Write([]byte(`{"id":"wf-1","name":"Workflow 1","enabled":true}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"wf-2","name":"Workflow 2","enabled":true}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"wf-2","name":"Updated Workflow","enabled":false}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	workflows, err := c.GetWorkflows()
	if err != nil || len(workflows) != 1 {
		t.Fatalf("GetWorkflows failed: %v", err)
	}

	wf, err := c.GetWorkflow("wf-1")
	if err != nil || wf.Name != "Workflow 1" {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	created, err := c.CreateWorkflow(CreateWorkflowRequest{Name: "Workflow 2", Enabled: true})
	if err != nil || created.ID != "wf-2" {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	enabled := false
	updated, err := c.UpdateWorkflow("wf-2", UpdateWorkflowRequest{Name: "Updated Workflow", Enabled: &enabled})
	if err != nil || updated.Name != "Updated Workflow" || updated.Enabled {
		t.Fatalf("UpdateWorkflow failed: %v", err)
	}

	err = c.DeleteWorkflow("wf-2")
	if err != nil {
		t.Fatalf("DeleteWorkflow failed: %v", err)
	}
}
