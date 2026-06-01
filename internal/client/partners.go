package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Partner struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	InTimeline       bool   `json:"inTimeline"`
	AvatarColor      string `json:"avatarColor"`
	ProfileImagePath string `json:"profileImagePath"`
}

type UpdatePartnerRequest struct {
	InTimeline bool `json:"inTimeline"`
}

func (c *Client) GetPartners() ([]Partner, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/partners", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var partners []Partner
	err = json.Unmarshal(body, &partners)
	if err != nil {
		return nil, err
	}

	return partners, nil
}

func (c *Client) CreatePartner(id string) (*Partner, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/partners/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var partner Partner
	err = json.Unmarshal(body, &partner)
	if err != nil {
		return nil, err
	}

	return &partner, nil
}

func (c *Client) UpdatePartner(id string, update UpdatePartnerRequest) (*Partner, error) {
	rb, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/partners/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var partner Partner
	err = json.Unmarshal(body, &partner)
	if err != nil {
		return nil, err
	}

	return &partner, nil
}

func (c *Client) DeletePartner(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/partners/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
