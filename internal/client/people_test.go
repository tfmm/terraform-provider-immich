package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestPeopleClientHTTPMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/people" {
				w.Write([]byte(`{"people":[{"id":"p-1","name":"Person 1"}]}`))
				return
			}
			if r.URL.Path == "/people/p-1" {
				w.Write([]byte(`{"id":"p-1","name":"Person 1"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"p-2","name":"Person 2"}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"p-2","name":"Updated Person"}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	people, err := c.GetPeople(true)
	if err != nil || len(people) != 1 {
		t.Fatalf("GetPeople failed: %v", err)
	}

	person, err := c.GetPerson("p-1")
	if err != nil || person.Name != "Person 1" {
		t.Fatalf("GetPerson failed: %v", err)
	}

	created, err := c.CreatePerson(CreatePersonRequest{Name: "Person 2"})
	if err != nil || created.ID != "p-2" {
		t.Fatalf("CreatePerson failed: %v", err)
	}

	updated, err := c.UpdatePerson("p-2", UpdatePersonRequest{Name: "Updated Person"})
	if err != nil || updated.Name != "Updated Person" {
		t.Fatalf("UpdatePerson failed: %v", err)
	}

	err = c.DeletePerson("p-2")
	if err != nil {
		t.Fatalf("DeletePerson failed: %v", err)
	}
}
