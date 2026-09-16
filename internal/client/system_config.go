package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SystemConfig struct {
	Backup           map[string]interface{} `json:"backup"`
	FFmpeg           map[string]interface{} `json:"ffmpeg"`
	Logging          map[string]interface{} `json:"logging"`
	MachineLearning  map[string]interface{} `json:"machineLearning"`
	Map              map[string]interface{} `json:"map"`
	NewVersionCheck  map[string]interface{} `json:"newVersionCheck"`
	NightlyTasks     map[string]interface{} `json:"nightlyTasks"`
	OAuth            map[string]interface{} `json:"oauth"`
	PasswordLogin    map[string]interface{} `json:"passwordLogin"`
	ReverseGeocoding map[string]interface{} `json:"reverseGeocoding"`
	Metadata         map[string]interface{} `json:"metadata"`
	StorageTemplate  map[string]interface{} `json:"storageTemplate"`
	Job              map[string]interface{} `json:"job"`
	Image            map[string]interface{} `json:"image"`
	Trash            map[string]interface{} `json:"trash"`
	Theme            map[string]interface{} `json:"theme"`
	Library          map[string]interface{} `json:"library"`
	Notifications    map[string]interface{} `json:"notifications"`
	Templates        map[string]interface{} `json:"templates"`
	Server           map[string]interface{} `json:"server"`
	User             map[string]interface{} `json:"user"`
}

func (c *Client) GetSystemConfig(ctx context.Context) (*SystemConfig, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/admin/config", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		req2, err2 := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/system-config", c.HostURL), nil)
		if err2 != nil {
			return nil, err
		}
		body, err = c.doRequest(req2)
		if err != nil {
			return nil, err
		}
	}

	var config SystemConfig
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Client) UpdateSystemConfig(ctx context.Context, config SystemConfig) (*SystemConfig, error) {
	rb, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/admin/config", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		req2, err2 := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/system-config", c.HostURL), bytes.NewBuffer(rb))
		if err2 != nil {
			return nil, err
		}
		body, err = c.doRequest(req2)
		if err != nil {
			return nil, err
		}
	}

	var updatedConfig SystemConfig
	err = json.Unmarshal(body, &updatedConfig)
	if err != nil {
		return nil, err
	}

	return &updatedConfig, nil
}
