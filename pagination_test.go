package supermarket

import (
	"net/url"
	"testing"
)

func TestPageHasMoreAndNextStart(t *testing.T) {
	p := Page[string]{Start: 0, Total: 10, Items: []string{"a", "b", "c"}}
	if !p.HasMore() {
		t.Error("HasMore should be true when items < total")
	}
	if got := p.NextStart(); got != 3 {
		t.Errorf("NextStart = %d, want 3", got)
	}

	last := Page[string]{Start: 7, Total: 10, Items: []string{"a", "b", "c"}}
	if last.HasMore() {
		t.Error("HasMore should be false when start+len == total")
	}
	if got := last.NextStart(); got != 0 {
		t.Errorf("NextStart on final page = %d, want 0", got)
	}
}

func TestPageEmptyResultsAreNotMore(t *testing.T) {
	p := Page[string]{Start: 0, Total: 0, Items: nil}
	if p.HasMore() {
		t.Error("HasMore on an empty page should be false")
	}
	if p.NextStart() != 0 {
		t.Errorf("NextStart on empty page = %d, want 0", p.NextStart())
	}
}

func TestPageHandlesOverflowGracefully(t *testing.T) {
	// Defensive: if a server returns more items than it advertises in
	// Total (shouldn't happen, but if it did) HasMore must be false.
	p := Page[string]{Start: 8, Total: 10, Items: []string{"a", "b", "c", "d"}}
	if p.HasMore() {
		t.Error("HasMore should be false when start+len exceeds total")
	}
}

func TestPageEmptyItemsDoesNotAdvance(t *testing.T) {
	// A server that reports Total greater than Start but returns no
	// items (a hiccup, or a Start past the real end) must not trick a
	// caller into an infinite pagination loop: HasMore is false and
	// NextStart does not return a non-advancing offset.
	p := Page[string]{Start: 0, Total: 10, Items: nil}
	if p.HasMore() {
		t.Error("HasMore must be false when the page carries no items")
	}
	if got := p.NextStart(); got != 0 {
		t.Errorf("NextStart on an empty page = %d, want 0 (no infinite loop)", got)
	}
}

func TestApplyPageQueryOmitsZeroValues(t *testing.T) {
	q := url.Values{}
	applyPageQuery(q, 0, 0)
	if len(q) != 0 {
		t.Errorf("applyPageQuery with zero start/items added params: %v", q)
	}

	q = url.Values{}
	applyPageQuery(q, 5, 10)
	if q.Get("start") != "5" || q.Get("items") != "10" {
		t.Errorf("applyPageQuery = %v", q)
	}

	// Only one of the two: applyPageQuery should skip the zero.
	q = url.Values{}
	applyPageQuery(q, 3, 0)
	if q.Get("start") != "3" {
		t.Error("expected start=3")
	}
	if q.Has("items") {
		t.Errorf("items should be absent when zero, got %q", q.Get("items"))
	}
}
