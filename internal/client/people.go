package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Person struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	BirthDate     string `json:"birthDate,omitempty"`
	ThumbnailPath string `json:"thumbnailPath,omitempty"`
	IsHidden      bool   `json:"isHidden"`
	IsFavorite    bool   `json:"isFavorite"`
}

type UpdatePersonRequest struct {
	Name       string `json:"name,omitempty"`
	BirthDate  string `json:"birthDate,omitempty"`
	IsHidden   *bool  `json:"isHidden,omitempty"`
	IsFavorite *bool  `json:"isFavorite,omitempty"`
}

func (c *Client) GetPeople(withHidden bool) ([]Person, error) {
	url := fmt.Sprintf("%s/people?withHidden=%v", c.HostURL, withHidden)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	// The API might return a paginated response or a simple array. 
	// Based on docs, it returns an array of PersonResponseDto or a paginated response.
	// Actually, the docs say GET /people returns PeopleResponseDto which has people: PersonResponseDto[]
	// Let's check.
	var response struct {
		People []Person `json:"people"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		// Fallback to array if it's not wrapped
		var people []Person
		if err2 := json.Unmarshal(body, &people); err2 == nil {
			return people, nil
		}
		return nil, err
	}

	return response.People, nil
}

func (c *Client) GetPerson(id string) (*Person, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/people/%s", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var person Person
	err = json.Unmarshal(body, &person)
	if err != nil {
		return nil, err
	}

	return &person, nil
}

func (c *Client) UpdatePerson(id string, person UpdatePersonRequest) (*Person, error) {
	rb, err := json.Marshal(person)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/people/%s", c.HostURL, id), bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var updatedPerson Person
	err = json.Unmarshal(body, &updatedPerson)
	if err != nil {
		return nil, err
	}

	return &updatedPerson, nil
}

func (c *Client) DeletePerson(id string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/people/%s", c.HostURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	return err
}
