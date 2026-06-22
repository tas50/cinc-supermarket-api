package supermarket

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUsersGetMapsNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/users/ghost", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error_messages":["no user"],"error_code":"NOT_FOUND"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	_, _, err := c.Users.Get(context.Background(), "ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUsersGetReturnsProfile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/users/alice", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"username":"alice","name":"Alice","company":"Acme",
			"github":["alice","alice2"],
			"cookbooks":{
				"owns":[{"cookbook_name":"apache2","cookbook":"u","cookbook_description":"","cookbook_maintainer":"alice"}],
				"collaborates":[],
				"follows":[]
			},
			"tools":{"owns":[]}
		}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	u, _, err := c.Users.Get(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Users.Get: %v", err)
	}
	if u.Username != "alice" || len(u.GitHub) != 2 {
		t.Errorf("u = %+v", u)
	}
	if len(u.Cookbooks.Owns) != 1 || u.Cookbooks.Owns[0].Name != "apache2" {
		t.Errorf("Owns = %+v", u.Cookbooks.Owns)
	}
}

func TestUsersGetExposesSlack(t *testing.T) {
	// The profile endpoint returns the contact handle under "slack"
	// (the old "irc" field is gone); make sure it decodes.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/users/alice", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"username":"alice","name":"Alice","twitter":"@alice","slack":"alice-slack",
			"github":[],"cookbooks":{"owns":[],"collaborates":[],"follows":[]},"tools":{"owns":[]}
		}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newTestClient(t, srv, false)

	u, _, err := c.Users.Get(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Users.Get: %v", err)
	}
	if u.Slack != "alice-slack" {
		t.Errorf("Slack = %q, want alice-slack", u.Slack)
	}
	if u.Twitter != "@alice" {
		t.Errorf("Twitter = %q", u.Twitter)
	}
}
