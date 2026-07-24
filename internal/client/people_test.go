package client

import (
	"encoding/json"
	"testing"
)

func TestPersonJSON(t *testing.T) {
	jsonInput := `{"id":"123","name":"John Doe","birthDate":"2023-06-10T00:00:00.000Z","isHidden":false,"isFavorite":true}`

	var p Person
	if err := json.Unmarshal([]byte(jsonInput), &p); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if p.BirthDate == nil {
		t.Fatalf("expected non-nil BirthDate")
	}

	if *p.BirthDate != "2023-06-10T00:00:00.000Z" {
		t.Errorf("expected '2023-06-10T00:00:00.000Z', got %q", *p.BirthDate)
	}
}

func TestUpdatePersonRequestJSON(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	req := UpdatePersonRequest{
		Name:      "John Doe",
		BirthDate: strPtr("1988-05-16"),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	expected := `{"name":"John Doe","birthDate":"1988-05-16"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}
