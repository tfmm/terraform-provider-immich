package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if r.URL.Path == "/admin/users" {
				w.Write([]byte(`[{"id":"usr-1","email":"user1@example.com","name":"User 1"}]`))
				return
			}
			if r.URL.Path == "/admin/users/usr-1" {
				w.Write([]byte(`{"id":"usr-1","email":"user1@example.com","name":"User 1"}`))
				return
			}
		case "POST":
			w.Write([]byte(`{"id":"usr-2","email":"user2@example.com","name":"User 2"}`))
			return
		case "PUT":
			w.Write([]byte(`{"id":"usr-2","email":"user2@example.com","name":"Updated User 2"}`))
			return
		case "DELETE":
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	users, err := c.GetUsers(context.Background())
	if err != nil || len(users) != 1 {
		t.Fatalf("GetUsers failed: %v", err)
	}

	usr, err := c.GetUser(context.Background(), "usr-1")
	if err != nil || usr.Name != "User 1" {
		t.Fatalf("GetUser failed: %v", err)
	}

	created, err := c.CreateUser(context.Background(), UserAdminCreateRequest{Email: "user2@example.com", Name: "User 2", Password: "pass"})
	if err != nil || created.ID != "usr-2" {
		t.Fatalf("CreateUser failed: %v", err)
	}

	updated, err := c.UpdateUser(context.Background(), "usr-2", UserAdminUpdateRequest{Name: "Updated User 2"})
	if err != nil || updated.Name != "Updated User 2" {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	err = c.DeleteUser(context.Background(), "usr-2")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}
