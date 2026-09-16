package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Asset struct {
	ID               string `json:"id"`
	OwnerId          string `json:"ownerId"`
	Type             string `json:"type"` // IMAGE or VIDEO
	OriginalFileName string `json:"originalFileName"`
	FileCreatedAt    string `json:"fileCreatedAt"`
	FileModifiedAt   string `json:"fileModifiedAt"`
	UpdatedAt        string `json:"updatedAt"`
	IsFavorite       bool   `json:"isFavorite"`
	IsArchived       bool   `json:"isArchived"`
	Description      string `json:"description"`
	Duration         string `json:"duration,omitempty"`
	ExifInfo         *Exif  `json:"exifInfo,omitempty"`
}

type Exif struct {
	Make             string  `json:"make,omitempty"`
	Model            string  `json:"model,omitempty"`
	ExifImageWidth   float64 `json:"exifImageWidth,omitempty"`
	ExifImageHeight  float64 `json:"exifImageHeight,omitempty"`
	DateTimeOriginal string  `json:"dateTimeOriginal,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
	City             string  `json:"city,omitempty"`
	State            string  `json:"state,omitempty"`
	Country          string  `json:"country,omitempty"`
}

type UpdateAssetRequest struct {
	IsFavorite  *bool    `json:"isFavorite,omitempty"`
	IsArchived  *bool    `json:"isArchived,omitempty"`
	Description string   `json:"description,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

type SearchAssetsRequest struct {
	IsFavorite       *bool    `json:"isFavorite,omitempty"`
	Type             string   `json:"type,omitempty"` // IMAGE or VIDEO
	OriginalFileName string   `json:"originalFileName,omitempty"`
	City             string   `json:"city,omitempty"`
	Country          string   `json:"country,omitempty"`
	Make             string   `json:"make,omitempty"`
	Model            string   `json:"model,omitempty"`
	WithExif         bool     `json:"withExif,omitempty"`
	Size             int      `json:"size,omitempty"`
	Page             int      `json:"page,omitempty"`
	AlbumIds         []string `json:"albumIds,omitempty"`
	PersonIds        []string `json:"personIds,omitempty"`
}

type SearchAssetsResponse struct {
	Assets struct {
		Total int     `json:"total"`
		Count int     `json:"count"`
		Items []Asset `json:"items"`
	} `json:"assets"`
}

func (c *Client) GetAsset(ctx context.Context, id string) (*Asset, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/assets/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var asset Asset
	err = json.Unmarshal(body, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (c *Client) UpdateAsset(ctx context.Context, id string, update UpdateAssetRequest) (*Asset, error) {
	rb, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/assets/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedAsset Asset
	err = json.Unmarshal(body, &updatedAsset)
	if err != nil {
		return nil, err
	}

	return &updatedAsset, nil
}

func (c *Client) DeleteAssets(ctx context.Context, ids []string) error {
	type DeleteRequest struct {
		Ids []string `json:"ids"`
	}

	rb, err := json.Marshal(DeleteRequest{Ids: ids})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/assets", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}

func (c *Client) SearchAssets(ctx context.Context, search SearchAssetsRequest) (*SearchAssetsResponse, error) {
	rb, err := json.Marshal(search)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/search/metadata", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response SearchAssetsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) UploadAsset(ctx context.Context, filePath string, fileCreatedAt, fileModifiedAt time.Time, isFavorite bool) (*Asset, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Stream the file into the request body instead of buffering it
	// entirely in memory first, since Immich assets (photos/videos) can be
	// large. multipart.Writer must be driven from a separate goroutine
	// since io.Pipe's writer blocks until the reader consumes the data.
	go func() {
		pw.CloseWithError(writeAssetUploadBody(writer, filePath, fileCreatedAt, fileModifiedAt, isFavorite))
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/assets", c.HostURL), pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resBody, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var asset Asset
	err = json.Unmarshal(resBody, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func writeAssetUploadBody(writer *multipart.Writer, filePath string, fileCreatedAt, fileModifiedAt time.Time, isFavorite bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	part, err := writer.CreateFormFile("assetData", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.WriteField("fileCreatedAt", fileCreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := writer.WriteField("fileModifiedAt", fileModifiedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	favorite := "false"
	if isFavorite {
		favorite = "true"
	}
	if err := writer.WriteField("isFavorite", favorite); err != nil {
		return err
	}

	return writer.Close()
}
