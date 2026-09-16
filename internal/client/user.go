package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID                   string `json:"id,omitempty"`
	Email                string `json:"email"`
	Name                 string `json:"name"`
	IsAdmin              bool   `json:"isAdmin"`
	StorageLabel         string `json:"storageLabel,omitempty"`
	QuotaSizeInBytes     *int64 `json:"quotaSizeInBytes,omitempty"`
	ShouldChangePassword bool   `json:"shouldChangePassword,omitempty"`
	CreatedAt            string `json:"createdAt,omitempty"`
	UpdatedAt            string `json:"updatedAt,omitempty"`
}

type UserAdminCreateRequest struct {
	Email                string `json:"email"`
	Password             string `json:"password,omitempty"`
	Name                 string `json:"name"`
	IsAdmin              bool   `json:"isAdmin,omitempty"`
	StorageLabel         string `json:"storageLabel,omitempty"`
	QuotaSizeInBytes     *int64 `json:"quotaSizeInBytes,omitempty"`
	ShouldChangePassword bool   `json:"shouldChangePassword,omitempty"`
}

type UserAdminUpdateRequest struct {
	Email                string `json:"email,omitempty"`
	Password             string `json:"password,omitempty"`
	Name                 string `json:"name,omitempty"`
	IsAdmin              bool   `json:"isAdmin,omitempty"`
	StorageLabel         string `json:"storageLabel,omitempty"`
	QuotaSizeInBytes     *int64 `json:"quotaSizeInBytes,omitempty"`
	ShouldChangePassword bool   `json:"shouldChangePassword,omitempty"`
}

func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/admin/users", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/admin/users/%s", c.HostURL, userID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) CreateUser(ctx context.Context, user UserAdminCreateRequest) (*User, error) {
	rb, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/admin/users", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newUser User
	err = json.Unmarshal(body, &newUser)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (c *Client) UpdateUser(ctx context.Context, userID string, user UserAdminUpdateRequest) (*User, error) {
	rb, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/admin/users/%s", c.HostURL, userID), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedUser User
	err = json.Unmarshal(body, &updatedUser)
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	// UserAdminDeleteDto has force: boolean
	// For simplicity, we'll force delete if needed, or just send empty object if it works.
	// Actually, the DTO says optional.
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/admin/users/%s", c.HostURL, userID), bytes.NewBufferString("{}"))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
