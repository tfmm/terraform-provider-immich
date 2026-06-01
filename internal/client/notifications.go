package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Notification struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Level       string                 `json:"level"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	CreatedAt   string                 `json:"createdAt"`
	ReadAt      *string                `json:"readAt"`
	Data        map[string]interface{} `json:"data"`
}

type CreateAdminNotificationRequest struct {
	Type        string                 `json:"type"` // SYSTEM
	Level       string                 `json:"level"` // INFO, WARNING, ERROR
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

func (c *Client) GetNotifications(unread bool) ([]Notification, error) {
	url := fmt.Sprintf("%s/notifications?unread=%v", c.HostURL, unread)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	err = json.Unmarshal(body, &notifications)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (c *Client) CreateAdminNotification(notification CreateAdminNotificationRequest) (*Notification, error) {
	rb, err := json.Marshal(notification)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/notifications", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newNotification Notification
	err = json.Unmarshal(body, &newNotification)
	if err != nil {
		return nil, err
	}

	return &newNotification, nil
}

func (c *Client) DeleteNotification(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notifications/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
