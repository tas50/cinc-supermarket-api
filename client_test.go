package supermarket

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
)

func TestNewClientDefaultsBaseURL(t *testing.T) {
	c, err := NewClient(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got := c.baseURL.String(); got != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", got, DefaultBaseURL)
	}
	if c.canSign() {
		t.Error("canSign should be false for an anonymous client")
	}
}

func TestNewClientRejectsInvalidBaseURL(t *testing.T) {
	if _, err := NewClient(Config{BaseURL: "not a url"}); err == nil {
		t.Error("expected an error for a malformed BaseURL")
	}
}

func TestCanSignRequiresBothUsernameAndKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	noUser, _ := NewClient(Config{Key: key})
	if noUser.canSign() {
		t.Error("canSign should be false without a Username")
	}

	noKey, _ := NewClient(Config{Username: "me"})
	if noKey.canSign() {
		t.Error("canSign should be false without a Key")
	}

	full, _ := NewClient(Config{Username: "me", Key: key})
	if !full.canSign() {
		t.Error("canSign should be true with both Username and Key")
	}
}

func TestParseKeyErrorMessagePrefixed(t *testing.T) {
	_, err := ParseKey([]byte("not a pem block"))
	if err == nil || !strings.Contains(err.Error(), "supermarket:") {
		t.Errorf("err = %v, want a supermarket-prefixed error", err)
	}
}
