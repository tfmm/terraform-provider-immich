package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Activity struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // COMMENT or LIKE
	AssetId   string `json:"assetId,omitempty"`
	AlbumId   string `json:"albumId"`
	User      User   `json:"user"`
	Comment   string `json:"comment,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type CreateActivityRequest struct {
	Type    string `json:"type"` // comment or like
	AlbumId string `json:"albumId"`
	AssetId string `json:"assetId,omitempty"`
	Comment string `json:"comment,omitempty"`
}

func (c *Client) GetActivities(albumId string, assetId string) ([]Activity, error) {
	url := fmt.Sprintf("%s/activities?albumId=%s", c.HostURL, albumId)
	if assetId != "" {
		url = fmt.Sprintf("%s&assetId=%s", url, assetId)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var activities []Activity
	err = json.Unmarshal(body, &activities)
	if err != nil {
		return nil, err
	}

	return activities, nil
}

func (c *Client) GetActivity(id string) (*Activity, error) {
	// Immich doesn't seem to have a direct GET /activities/{id} based on research.
	// We might have to find it in the list if we want to "Read" it by ID alone.
	// But usually Terraform Read has the ID. 
	// If the API doesn't support GET by ID, we'll have to return an error or handle it.
	// Wait, some research showed DELETE /activities/{id}.
	// If there's no GET /activities/{id}, we might have a problem with pure Terraform Read.
	// However, we can probably use a data source or just return the state if we can't refresh it easily.
	// Actually, let's assume it doesn't exist for now and see if we can find a workaround.
	return nil, fmt.Errorf("GET /activities/{id} is not supported by Immich API")
}

func (c *Client) CreateActivity(activity CreateActivityRequest) (*Activity, error) {
	rb, err := json.Marshal(activity)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/activities", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newActivity Activity
	err = json.Unmarshal(body, &newActivity)
	if err != nil {
		return nil, err
	}

	return &newActivity, nil
}

func (c *Client) DeleteActivity(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/activities/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
