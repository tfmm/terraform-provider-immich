package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AlbumUser struct {
	User *User  `json:"user,omitempty"`
	Role string `json:"role"`
}

type Album struct {
	ID                    string      `json:"id,omitempty"`
	AlbumName             string      `json:"albumName"`
	Description           string      `json:"description"`
	CreatedAt             string      `json:"createdAt,omitempty"`
	UpdatedAt             string      `json:"updatedAt,omitempty"`
	AlbumThumbnailAssetId *string     `json:"albumThumbnailAssetId"`
	Shared                bool        `json:"shared"`
	AlbumUsers            []AlbumUser `json:"albumUsers"`
	HasSharedLink         bool        `json:"hasSharedLink"`
	AssetCount            int         `json:"assetCount"`
	IsActivityEnabled     bool        `json:"isActivityEnabled"`
	Order                 string      `json:"order,omitempty"`
}

type AlbumUserCreate struct {
	UserId string `json:"userId"`
	Role   string `json:"role"`
}

type CreateAlbumRequest struct {
	AlbumName   string            `json:"albumName"`
	Description string            `json:"description,omitempty"`
	AlbumUsers  []AlbumUserCreate `json:"albumUsers,omitempty"`
	AssetIds    []string          `json:"assetIds,omitempty"`
}

type UpdateAlbumRequest struct {
	AlbumName             string  `json:"albumName,omitempty"`
	Description           string  `json:"description,omitempty"`
	AlbumThumbnailAssetId *string `json:"albumThumbnailAssetId,omitempty"`
	IsActivityEnabled     *bool   `json:"isActivityEnabled,omitempty"`
	Order                 string  `json:"order,omitempty"`
}

type BulkIdsRequest struct {
	Ids []string `json:"ids"`
}

func (c *Client) GetAlbums() ([]Album, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/albums", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var albums []Album
	err = json.Unmarshal(body, &albums)
	if err != nil {
		return nil, err
	}

	return albums, nil
}

func (c *Client) GetAlbum(id string) (*Album, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/albums/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var album Album
	err = json.Unmarshal(body, &album)
	if err != nil {
		return nil, err
	}

	return &album, nil
}

func (c *Client) CreateAlbum(data CreateAlbumRequest) (*Album, error) {
	rb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/albums", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var album Album
	err = json.Unmarshal(body, &album)
	if err != nil {
		return nil, err
	}

	return &album, nil
}

func (c *Client) UpdateAlbum(id string, data UpdateAlbumRequest) (*Album, error) {
	rb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/albums/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var album Album
	err = json.Unmarshal(body, &album)
	if err != nil {
		return nil, err
	}

	return &album, nil
}

func (c *Client) DeleteAlbum(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/albums/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

func (c *Client) AddAssetsToAlbum(albumId string, assetIds []string) error {
	data := BulkIdsRequest{Ids: assetIds}
	rb, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/albums/%s/assets", c.HostURL, albumId), bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

type AddUsersRequest struct {
	AlbumUsers []AlbumUserCreate `json:"albumUsers"`
}

type UpdateAlbumUserRequest struct {
	Role string `json:"role"`
}

func (c *Client) AddUsersToAlbum(albumId string, users []AlbumUserCreate) (*Album, error) {
	data := AddUsersRequest{AlbumUsers: users}
	rb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/albums/%s/users", c.HostURL, albumId), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var album Album
	err = json.Unmarshal(body, &album)
	if err != nil {
		return nil, err
	}

	return &album, nil
}

func (c *Client) UpdateAlbumUserRole(albumId string, userId string, role string) (*Album, error) {
	data := UpdateAlbumUserRequest{Role: role}
	rb, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/albums/%s/user/%s", c.HostURL, albumId, userId), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var album Album
	err = json.Unmarshal(body, &album)
	if err != nil {
		return nil, err
	}

	return &album, nil
}

func (c *Client) RemoveUserFromAlbum(albumId string, userId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/albums/%s/user/%s", c.HostURL, albumId, userId), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

