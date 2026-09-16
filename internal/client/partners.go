package client

import (
	"bytes"
	"context"
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

func (c *Client) GetPartners(ctx context.Context) ([]Partner, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/partners", c.HostURL), nil)
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

type CreatePartnerRequest struct {
	SharedWithId string `json:"sharedWithId"`
}

func (c *Client) CreatePartner(ctx context.Context, id string) (*Partner, error) {
	rb, err := json.Marshal(CreatePartnerRequest{SharedWithId: id})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/partners", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		req2, err2 := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/partners/%s", c.HostURL, id), nil)
		if err2 == nil {
			if body2, err3 := c.doRequest(req2); err3 == nil {
				body = body2
				err = nil
			}
		}
		if err != nil {
			return nil, err
		}
	}

	var partner Partner
	err = json.Unmarshal(body, &partner)
	if err != nil {
		return nil, err
	}

	return &partner, nil
}

func (c *Client) UpdatePartner(ctx context.Context, id string, update UpdatePartnerRequest) (*Partner, error) {
	rb, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/partners/%s", c.HostURL, id), bytes.NewBuffer(rb))
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

func (c *Client) DeletePartner(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/partners/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
