package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Workflow struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	// Triggers, Filters, Actions would be complex nested objects.
	// For simplicity in this experimental implementation, we'll use raw maps.
	Triggers []map[string]interface{} `json:"triggers"`
	Filters  []map[string]interface{} `json:"filters"`
	Actions  []map[string]interface{} `json:"actions"`
}

type CreateWorkflowRequest struct {
	Name     string                   `json:"name"`
	Enabled  bool                     `json:"enabled"`
	Triggers []map[string]interface{} `json:"triggers"`
	Filters  []map[string]interface{} `json:"filters"`
	Actions  []map[string]interface{} `json:"actions"`
}

type UpdateWorkflowRequest struct {
	Name     string                   `json:"name,omitempty"`
	Enabled  *bool                    `json:"enabled,omitempty"`
	Triggers []map[string]interface{} `json:"triggers,omitempty"`
	Filters  []map[string]interface{} `json:"filters,omitempty"`
	Actions  []map[string]interface{} `json:"actions,omitempty"`
}

func (c *Client) GetWorkflows() ([]Workflow, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/workflow", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var workflows []Workflow
	err = json.Unmarshal(body, &workflows)
	if err != nil {
		return nil, err
	}

	return workflows, nil
}

func (c *Client) GetWorkflow(id string) (*Workflow, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/workflow/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var workflow Workflow
	err = json.Unmarshal(body, &workflow)
	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

func (c *Client) CreateWorkflow(workflow CreateWorkflowRequest) (*Workflow, error) {
	rb, err := json.Marshal(workflow)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/workflow", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newWorkflow Workflow
	err = json.Unmarshal(body, &newWorkflow)
	if err != nil {
		return nil, err
	}

	return &newWorkflow, nil
}

func (c *Client) UpdateWorkflow(id string, workflow UpdateWorkflowRequest) (*Workflow, error) {
	rb, err := json.Marshal(workflow)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/workflow/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedWorkflow Workflow
	err = json.Unmarshal(body, &updatedWorkflow)
	if err != nil {
		return nil, err
	}

	return &updatedWorkflow, nil
}

func (c *Client) DeleteWorkflow(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/workflow/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
