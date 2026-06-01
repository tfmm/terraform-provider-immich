package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Library struct {
	ID                string   `json:"id"`
	OwnerId           string   `json:"ownerId"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	ImportPaths       []string `json:"importPaths"`
	ExclusionPatterns []string `json:"exclusionPatterns"`
	AssetCount        int      `json:"assetCount"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
	RefreshedAt       string   `json:"refreshedAt"`
}

type CreateLibraryRequest struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	ImportPaths       []string `json:"importPaths"`
	ExclusionPatterns []string `json:"exclusionPatterns"`
	IsVisible         bool     `json:"isVisible"`
}

type UpdateLibraryRequest struct {
	Name              string   `json:"name,omitempty"`
	ImportPaths       []string `json:"importPaths,omitempty"`
	ExclusionPatterns []string `json:"exclusionPatterns,omitempty"`
	IsVisible         *bool    `json:"isVisible,omitempty"`
}

func (c *Client) GetLibraries() ([]Library, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/library", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var libraries []Library
	err = json.Unmarshal(body, &libraries)
	if err != nil {
		return nil, err
	}

	return libraries, nil
}

func (c *Client) GetLibrary(id string) (*Library, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/library/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var library Library
	err = json.Unmarshal(body, &library)
	if err != nil {
		return nil, err
	}

	return &library, nil
}

func (c *Client) CreateLibrary(library CreateLibraryRequest) (*Library, error) {
	rb, err := json.Marshal(library)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/library", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newLibrary Library
	err = json.Unmarshal(body, &newLibrary)
	if err != nil {
		return nil, err
	}

	return &newLibrary, nil
}

func (c *Client) UpdateLibrary(id string, library UpdateLibraryRequest) (*Library, error) {
	rb, err := json.Marshal(library)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/library/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedLibrary Library
	err = json.Unmarshal(body, &updatedLibrary)
	if err != nil {
		return nil, err
	}

	return &updatedLibrary, nil
}

func (c *Client) DeleteLibrary(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/library/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
