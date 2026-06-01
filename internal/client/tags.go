package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Tag struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // OBJECT or USER
	// Color  string `json:"color,omitempty"`
}

type CreateTagRequest struct {
	Name string `json:"name"`
	Type string `json:"type"` // OBJECT or USER
}

type UpdateTagRequest struct {
	Name string `json:"name"`
}

func (c *Client) GetTags() ([]Tag, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tag", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var tags []Tag
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (c *Client) GetTag(id string) (*Tag, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tag/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var tag Tag
	err = json.Unmarshal(body, &tag)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func (c *Client) CreateTag(tag CreateTagRequest) (*Tag, error) {
	rb, err := json.Marshal(tag)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/tag", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newTag Tag
	err = json.Unmarshal(body, &newTag)
	if err != nil {
		return nil, err
	}

	return &newTag, nil
}

func (c *Client) UpdateTag(id string, tag UpdateTagRequest) (*Tag, error) {
	rb, err := json.Marshal(tag)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/tag/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedTag Tag
	err = json.Unmarshal(body, &updatedTag)
	if err != nil {
		return nil, err
	}

	return &updatedTag, nil
}

func (c *Client) DeleteTag(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/tag/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
