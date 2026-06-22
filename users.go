package supermarket

import (
	"context"
	"net/http"
	"net/url"
)

// User is the profile representation returned by GET /users/:username.
type User struct {
	Username  string        `json:"username"`
	Name      string        `json:"name"`
	Company   string        `json:"company"`
	GitHub    []string      `json:"github"`
	Twitter   string        `json:"twitter"`
	Slack     string        `json:"slack"`
	Cookbooks UserCookbooks `json:"cookbooks"`
	Tools     UserTools     `json:"tools"`
}

// UserCookbooks groups the cookbook relationships on a user profile.
type UserCookbooks struct {
	Owns         []CookbookSummary `json:"owns"`
	Collaborates []CookbookSummary `json:"collaborates"`
	Follows      []CookbookSummary `json:"follows"`
}

// UserTools groups the tool relationships on a user profile.
type UserTools struct {
	Owns []ToolSummary `json:"owns"`
}

// UsersService accesses the /users endpoints.
type UsersService struct{ client *Client }

// Get returns a user profile by username.
func (s *UsersService) Get(ctx context.Context, username string) (*User, *Response, error) {
	u, resp, err := doJSON[User](ctx, s.client, request{
		method: http.MethodGet,
		path:   "/api/v1/users/" + url.PathEscape(username),
	})
	if err != nil {
		return nil, resp, err
	}
	return &u, resp, nil
}
