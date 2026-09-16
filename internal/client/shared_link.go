package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SharedLink struct {
	ID            string  `json:"id"`
	Description   *string `json:"description"`
	UserId        string  `json:"userId"`
	Key           string  `json:"key"`
	Type          string  `json:"type"`
	CreatedAt     string  `json:"createdAt"`
	ExpiresAt     *string `json:"expiresAt"`
	AllowUpload   bool    `json:"allowUpload"`
	AllowDownload bool    `json:"allowDownload"`
	ShowMetadata  bool    `json:"showMetadata"`
	Slug          *string `json:"slug"`
}

type SharedLinkCreateRequest struct {
	Type          string   `json:"type"`
	AssetIds      []string `json:"assetIds,omitempty"`
	AlbumId       *string  `json:"albumId,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Password      *string  `json:"password,omitempty"`
	Slug          *string  `json:"slug,omitempty"`
	ExpiresAt     *string  `json:"expiresAt,omitempty"`
	AllowUpload   *bool    `json:"allowUpload,omitempty"`
	AllowDownload *bool    `json:"allowDownload,omitempty"`
	ShowMetadata  *bool    `json:"showMetadata,omitempty"`
}

type SharedLinkUpdateRequest struct {
	Description   *string `json:"description,omitempty"`
	Password      *string `json:"password,omitempty"`
	Slug          *string `json:"slug,omitempty"`
	ExpiresAt     *string `json:"expiresAt,omitempty"`
	AllowUpload   *bool   `json:"allowUpload,omitempty"`
	AllowDownload *bool   `json:"allowDownload,omitempty"`
	ShowMetadata  *bool   `json:"showMetadata,omitempty"`
}

func (c *Client) GetSharedLinks(ctx context.Context) ([]SharedLink, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/shared-links", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var sharedLinks []SharedLink
	err = json.Unmarshal(body, &sharedLinks)
	if err != nil {
		return nil, err
	}

	return sharedLinks, nil
}

func (c *Client) GetSharedLink(ctx context.Context, id string) (*SharedLink, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/shared-links/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var sharedLink SharedLink
	err = json.Unmarshal(body, &sharedLink)
	if err != nil {
		return nil, err
	}

	return &sharedLink, nil
}

func (c *Client) CreateSharedLink(ctx context.Context, data SharedLinkCreateRequest) (*SharedLink, error) {
	rb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/shared-links", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var sharedLink SharedLink
	err = json.Unmarshal(body, &sharedLink)
	if err != nil {
		return nil, err
	}

	return &sharedLink, nil
}

func (c *Client) UpdateSharedLink(ctx context.Context, id string, data SharedLinkUpdateRequest) (*SharedLink, error) {
	rb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/shared-links/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var sharedLink SharedLink
	err = json.Unmarshal(body, &sharedLink)
	if err != nil {
		return nil, err
	}

	return &sharedLink, nil
}

func (c *Client) DeleteSharedLink(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/shared-links/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
