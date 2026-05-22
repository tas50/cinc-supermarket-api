package supermarket

import "testing"

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
