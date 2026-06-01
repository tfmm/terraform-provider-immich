package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Stack struct {
	ID             string `json:"id"`
	PrimaryAssetId string `json:"primaryAssetId"`
	// Assets []AssetResponseDto `json:"assets"`
}

type CreateStackRequest struct {
	AssetIds []string `json:"assetIds"`
}

type UpdateStackRequest struct {
	PrimaryAssetId string `json:"primaryAssetId"`
}

func (c *Client) GetStacks() ([]Stack, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/stacks", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var stacks []Stack
	err = json.Unmarshal(body, &stacks)
	if err != nil {
		return nil, err
	}

	return stacks, nil
}

func (c *Client) GetStack(id string) (*Stack, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/stacks/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var stack Stack
	err = json.Unmarshal(body, &stack)
	if err != nil {
		return nil, err
	}

	return &stack, nil
}

func (c *Client) CreateStack(stack CreateStackRequest) (*Stack, error) {
	rb, err := json.Marshal(stack)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/stacks", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newStack Stack
	err = json.Unmarshal(body, &newStack)
	if err != nil {
		return nil, err
	}

	return &newStack, nil
}

func (c *Client) UpdateStack(id string, stack UpdateStackRequest) (*Stack, error) {
	rb, err := json.Marshal(stack)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/stacks/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedStack Stack
	err = json.Unmarshal(body, &updatedStack)
	if err != nil {
		return nil, err
	}

	return &updatedStack, nil
}

func (c *Client) DeleteStack(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/stacks/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
