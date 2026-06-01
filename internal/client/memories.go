package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Memory struct {
	ID       string `json:"id"`
	MemoryAt string `json:"memoryAt"`
	IsSaved  bool   `json:"isSaved"`
	// Assets []AssetResponseDto `json:"assets"`
}

type CreateMemoryRequest struct {
	MemoryAt string `json:"memoryAt"`
	IsSaved  bool   `json:"isSaved,omitempty"`
}

type UpdateMemoryRequest struct {
	IsSaved *bool `json:"isSaved,omitempty"`
}

func (c *Client) GetMemories() ([]Memory, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/memories", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var memories []Memory
	err = json.Unmarshal(body, &memories)
	if err != nil {
		return nil, err
	}

	return memories, nil
}

func (c *Client) GetMemory(id string) (*Memory, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/memories/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var memory Memory
	err = json.Unmarshal(body, &memory)
	if err != nil {
		return nil, err
	}

	return &memory, nil
}

func (c *Client) CreateMemory(memory CreateMemoryRequest) (*Memory, error) {
	rb, err := json.Marshal(memory)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/memories", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newMemory Memory
	err = json.Unmarshal(body, &newMemory)
	if err != nil {
		return nil, err
	}

	return &newMemory, nil
}

func (c *Client) UpdateMemory(id string, memory UpdateMemoryRequest) (*Memory, error) {
	rb, err := json.Marshal(memory)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/memories/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedMemory Memory
	err = json.Unmarshal(body, &updatedMemory)
	if err != nil {
		return nil, err
	}

	return &updatedMemory, nil
}

func (c *Client) DeleteMemory(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/memories/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
