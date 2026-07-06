package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Face struct {
	ID            string  `json:"id"`
	AssetId       string  `json:"assetId"`
	PersonId      string  `json:"personId"`
	BoundingBoxX1 float64 `json:"boundingBoxX1"`
	BoundingBoxY1 float64 `json:"boundingBoxY1"`
	BoundingBoxX2 float64 `json:"boundingBoxX2"`
	BoundingBoxY2 float64 `json:"boundingBoxY2"`
	ImageHeight   int     `json:"imageHeight"`
	ImageWidth    int     `json:"imageWidth"`
}

type CreateFaceRequest struct {
	AssetId       string  `json:"assetId"`
	PersonId      string  `json:"personId"`
	BoundingBoxX1 float64 `json:"boundingBoxX1"`
	BoundingBoxY1 float64 `json:"boundingBoxY1"`
	BoundingBoxX2 float64 `json:"boundingBoxX2"`
	BoundingBoxY2 float64 `json:"boundingBoxY2"`
	ImageHeight   int     `json:"imageHeight"`
	ImageWidth    int     `json:"imageWidth"`
}

type UpdateFaceRequest struct {
	PersonId string `json:"personId"`
}

func (c *Client) GetFaces(assetId string) ([]Face, error) {
	url := fmt.Sprintf("%s/faces?id=%s", c.HostURL, assetId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var faces []Face
	err = json.Unmarshal(body, &faces)
	if err != nil {
		return nil, err
	}

	return faces, nil
}

func (c *Client) CreateFace(face CreateFaceRequest) (*Face, error) {
	rb, err := json.Marshal(face)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/faces", c.HostURL), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var newFace Face
	err = json.Unmarshal(body, &newFace)
	if err != nil {
		return nil, err
	}

	return &newFace, nil
}

func (c *Client) UpdateFace(id string, update UpdateFaceRequest) (*Face, error) {
	type FaceDto struct {
		Id string `json:"id"`
	}

	rb, err := json.Marshal(FaceDto{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/faces/%s", c.HostURL, update.PersonId), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return &Face{ID: id, PersonId: update.PersonId}, nil
}

func (c *Client) DeleteFace(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/faces/%s", c.HostURL, id), bytes.NewBufferString(`{"force": true}`))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
