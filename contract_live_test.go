//go:build contract

// Package supermarket_test's live contract suite. It is intentionally OUT of
// the default `go test ./...` run (the `contract` build tag gates it) because
// it makes real network calls to the public Supermarket. Run it with:
//
//	go test -tags contract ./...
//
// It exercises the anonymous READ endpoints through the actual client and
// decoders and asserts our implementation still conforms to the live API: if
// Supermarket renames, removes, or retypes a field our client depends on, the
// matching test fails and names the drift. It is the drift alarm; the
// always-on replay suite (contract_replay_test.go) is its deterministic twin.
//
// Coverage boundary: this suite covers anonymous reads only. The signed WRITE
// path (share/delete) is guarded deterministically by the offline version
// guard and the mixlib verifier test in internal/signing; a true live signed
// upload needs a Supermarket account or a local instance and is future work.
package supermarket_test

import (
	"context"
	"testing"
	"time"

	supermarket "github.com/tas50/cinc-supermarket-api"
)

func liveClient(t *testing.T) (*supermarket.Client, context.Context, context.CancelFunc) {
	t.Helper()
	c, err := supermarket.NewClient(supermarket.Config{})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	return c, ctx, cancel
}

func TestContractLive_CookbooksListAndPagination(t *testing.T) {
	c, ctx, cancel := liveClient(t)
	defer cancel()

	page1, _, err := c.Cookbooks.List(ctx, supermarket.ListOptions{Items: 5})
	if err != nil {
		t.Fatalf("Cookbooks.List page 1: %v", err)
	}
	assertCookbookPage(t, page1, "live cookbooks list page 1")

	if !page1.HasMore() {
		t.Fatalf("expected the catalog to have more than 5 cookbooks; HasMore() = false (total=%d)", page1.Total)
	}
	page2, _, err := c.Cookbooks.List(ctx, supermarket.ListOptions{Start: page1.NextStart(), Items: 5})
	if err != nil {
		t.Fatalf("Cookbooks.List page 2: %v", err)
	}
	assertCookbookPage(t, page2, "live cookbooks list page 2")
	if page2.Start != page1.NextStart() {
		t.Errorf("page 2 Start = %d, want %d (the offset we requested)", page2.Start, page1.NextStart())
	}
	if page2.Start == page1.Start {
		t.Errorf("pagination did not advance: page 2 Start == page 1 Start (%d)", page1.Start)
	}
}

func TestContractLive_Search(t *testing.T) {
	c, ctx, cancel := liveClient(t)
	defer cancel()

	page, _, err := c.Search.Cookbooks(ctx, supermarket.SearchOptions{Q: "nginx", Items: 5})
	if err != nil {
		t.Fatalf("Search.Cookbooks: %v", err)
	}
	assertCookbookPage(t, page, "live search q=nginx")
}

func TestContractLive_CookbookShowAndVersion(t *testing.T) {
	c, ctx, cancel := liveClient(t)
	defer cancel()

	cb, _, err := c.Cookbooks.Get(ctx, "nginx")
	if err != nil {
		t.Fatalf("Cookbooks.Get(nginx): %v", err)
	}
	assertCookbook(t, cb, "nginx")

	// A specific, long-published version so the assertions don't chase a
	// moving "latest".
	v, _, err := c.Cookbooks.GetVersion(ctx, "nginx", "12.3.1")
	if err != nil {
		t.Fatalf("Cookbooks.GetVersion(nginx, 12.3.1): %v", err)
	}
	assertCookbookVersion(t, v, "12.3.1")
}

func TestContractLive_ToolsList(t *testing.T) {
	c, ctx, cancel := liveClient(t)
	defer cancel()

	page, _, err := c.Tools.List(ctx, supermarket.ToolsListOptions{Items: 5})
	if err != nil {
		t.Fatalf("Tools.List: %v", err)
	}
	assertToolPage(t, page, "live tools list")
}

func TestContractLive_Universe(t *testing.T) {
	c, ctx, cancel := liveClient(t)
	defer cancel()

	u, _, err := c.Universe.Get(ctx)
	if err != nil {
		t.Fatalf("Universe.Get: %v", err)
	}
	assertUniverse(t, u)
}

func TestContractLive_User(t *testing.T) {
	c, ctx, cancel := liveClient(t)
	defer cancel()

	// sous-chefs is the long-lived org account behind a large slice of the
	// catalog — a stable fixture whose profile shape we can pin.
	u, _, err := c.Users.Get(ctx, "sous-chefs")
	if err != nil {
		t.Fatalf("Users.Get(sous-chefs): %v", err)
	}
	assertUser(t, u, "sous-chefs")
}
