package supermarket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
