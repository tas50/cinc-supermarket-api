package supermarket_test

// Replay half of the contract layer: decode the recorded golden bodies under
// testdata/contract/ through our public types and run the shared assertions.
// This always runs in the default `go test ./...` — no network, no tags — so
// CI gets a deterministic signal that the decoders still accept real
// Supermarket payloads. The golden bodies were captured from
// supermarket.chef.io; refresh them by re-running the live contract suite's
// recordings (see README "Contract testing").

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	supermarket "github.com/tas50/cinc-supermarket-api"
)

func readGolden[T any](t *testing.T, name string) T {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "contract", name))
	if err != nil {
		t.Fatalf("reading golden %s: %v", name, err)
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decoding golden %s through public type %T: %v", name, out, err)
	}
	return out
}

func TestContractReplay_CookbooksList(t *testing.T) {
	p := readGolden[supermarket.Page[supermarket.CookbookSummary]](t, "cookbooks_list.json")
	assertCookbookPage(t, p, "cookbooks_list.json")
}

func TestContractReplay_Search(t *testing.T) {
	p := readGolden[supermarket.Page[supermarket.CookbookSummary]](t, "search_nginx.json")
	assertCookbookPage(t, p, "search_nginx.json")
}

func TestContractReplay_Cookbook(t *testing.T) {
	cb := readGolden[supermarket.Cookbook](t, "cookbook_nginx.json")
	assertCookbook(t, &cb, "nginx")
}

func TestContractReplay_CookbookVersion(t *testing.T) {
	v := readGolden[supermarket.CookbookVersion](t, "cookbook_version_nginx_12_3_1.json")
	assertCookbookVersion(t, &v, "12.3.1")
}

func TestContractReplay_ToolsList(t *testing.T) {
	p := readGolden[supermarket.Page[supermarket.ToolSummary]](t, "tools_list.json")
	assertToolPage(t, p, "tools_list.json")
}

func TestContractReplay_User(t *testing.T) {
	u := readGolden[supermarket.User](t, "user_sous-chefs.json")
	assertUser(t, &u, "sous-chefs")
}
