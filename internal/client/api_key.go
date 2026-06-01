package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ApiKey struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Permissions []string `json:"permissions"`
}

type ApiKeyCreateResponse struct {
	Secret string `json:"secret"`
	ApiKey ApiKey `json:"apiKey"`
}

type ApiKeyCreateRequest struct {
	Name        string   `json:"name,omitempty"`
	Permissions []string `json:"permissions"`
}

type ApiKeyUpdateRequest struct {
	Name        string   `json:"name,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

func (c *Client) GetApiKeys() ([]ApiKey, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api-keys", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var apiKeys []ApiKey
	err = json.Unmarshal(body, &apiKeys)
	if err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (c *Client) GetApiKey(apiKeyID string) (*ApiKey, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api-keys/%s", c.HostURL, apiKeyID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var apiKey ApiKey
	err = json.Unmarshal(body, &apiKey)
	if err != nil {
		return nil, err
	}

	return &apiKey, nil
}

func (c *Client) CreateApiKey(apiKey ApiKeyCreateRequest) (*ApiKeyCreateResponse, error) {
	rb, err := json.Marshal(apiKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api-keys", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var resp ApiKeyCreateResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) UpdateApiKey(apiKeyID string, apiKey ApiKeyUpdateRequest) (*ApiKey, error) {
	rb, err := json.Marshal(apiKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api-keys/%s", c.HostURL, apiKeyID), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedApiKey ApiKey
	err = json.Unmarshal(body, &updatedApiKey)
	if err != nil {
		return nil, err
	}

	return &updatedApiKey, nil
}

func (c *Client) DeleteApiKey(apiKeyID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api-keys/%s", c.HostURL, apiKeyID), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
