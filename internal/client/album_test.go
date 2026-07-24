package client

import (
	"encoding/json"
	"testing"
)

func TestAlbumJSONUnmarshal(t *testing.T) {
	jsonInput := `{"id":"album-1","albumName":"Test Album","description":null,"isActivityEnabled":true}`

	var a Album
	if err := json.Unmarshal([]byte(jsonInput), &a); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if a.Description != nil {
		t.Errorf("expected nil Description for null json, got %v", *a.Description)
	}
}

func TestCreateAlbumRequestJSON(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	req := CreateAlbumRequest{
		AlbumName:   "My Album",
		Description: strPtr("Album description"),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	expected := `{"albumName":"My Album","description":"Album description"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}
