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
		// Supermarket maps each owned/followed cookbook name to its API
		// URL (see UserCookbooks), and omits a relationship key when it's
		// empty — so "follows" is absent here and must decode to nil.
		_, _ = io.WriteString(w, `{
			"username":"alice","name":"Alice","company":"Acme",
			"github":["alice","alice2"],
			"cookbooks":{
				"owns":{"apache2":"https://supermarket.chef.io/api/v1/cookbooks/apache2"},
				"collaborates":{}
			},
			"tools":{"owns":{"knife-foo":"https://supermarket.chef.io/api/v1/tools/knife-foo"}}
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
	if got := u.Cookbooks.Owns["apache2"]; got != "https://supermarket.chef.io/api/v1/cookbooks/apache2" {
		t.Errorf("Owns[apache2] = %q, want the cookbook URL", got)
	}
	if u.Cookbooks.Follows != nil {
		t.Errorf("Follows should decode to nil when the key is absent, got %+v", u.Cookbooks.Follows)
	}
	if got := u.Tools.Owns["knife-foo"]; got != "https://supermarket.chef.io/api/v1/tools/knife-foo" {
		t.Errorf("Tools.Owns[knife-foo] = %q, want the tool URL", got)
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
			"github":[],"cookbooks":{"owns":{},"collaborates":{},"follows":{}},"tools":{"owns":{}}
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
