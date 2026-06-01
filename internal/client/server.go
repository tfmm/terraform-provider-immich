package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ServerAbout struct {
	Version     string `json:"version"`
	Build       string `json:"build"`
	NodeJS      string `json:"nodejs"`
	FFmpeg      string `json:"ffmpeg"`
	ExifTool    string `json:"exiftool"`
	ImageMagick string `json:"imagemagick"`
	Libvips     string `json:"libvips"`
}

type ServerFeatures struct {
	ConfigFile        bool `json:"configFile"`
	FacialRecognition bool `json:"facialRecognition"`
	Map               bool `json:"map"`
	ReverseGeocoding  bool `json:"reverseGeocoding"`
	Search            bool `json:"search"`
	Oauth             bool `json:"oauth"`
	PasswordLogin     bool `json:"passwordLogin"`
}

type ServerStatistics struct {
	Photos   int   `json:"photos"`
	Videos   int   `json:"videos"`
	Usage    int64 `json:"usage"`
	Users    int   `json:"users"`
}

func (c *Client) GetServerAbout() (*ServerAbout, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/about", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var about ServerAbout
	err = json.Unmarshal(body, &about)
	if err != nil {
		return nil, err
	}

	return &about, nil
}

func (c *Client) GetServerFeatures() (*ServerFeatures, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/features", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var features ServerFeatures
	err = json.Unmarshal(body, &features)
	if err != nil {
		return nil, err
	}

	return &features, nil
}

func (c *Client) GetServerStatistics() (*ServerStatistics, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/server/statistics", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var stats ServerStatistics
	err = json.Unmarshal(body, &stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
