package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationsClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`[{"id":"notif-1","title":"System Notice"}]`))
		case "POST":
			w.Write([]byte(`{"id":"notif-2","title":"Alert"}`))
		case "DELETE":
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "token")

	notifs, err := c.GetNotifications(true)
	if err != nil || len(notifs) != 1 {
		t.Fatalf("GetNotifications failed: %v", err)
	}

	created, err := c.CreateAdminNotification(CreateAdminNotificationRequest{Title: "Alert", Type: "SYSTEM", Level: "INFO"})
	if err != nil || created.ID != "notif-2" {
		t.Fatalf("CreateAdminNotification failed: %v", err)
	}

	err = c.DeleteNotification("notif-2")
	if err != nil {
		t.Fatalf("DeleteNotification failed: %v", err)
	}
}
