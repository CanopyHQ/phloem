package causal

import (
	"reflect"
	"strings"
	"testing"
)

func TestExtract_empty(t *testing.T) {
	out := Extract("")
	if out != nil {
		t.Errorf("Extract(\"\") want nil, got %v", out)
	}
	out = Extract("   ")
	if out != nil {
		t.Errorf("Extract(whitespace) want nil, got %v", out)
	}
}

func TestExtract_because(t *testing.T) {
	content := "We fixed the bug because the auth middleware was missing."
	out := Extract(content)
	if len(out) == 0 {
		t.Fatal("expected at least one relation")
	}
	found := false
	for _, r := range out {
		if r.Phrase != "" && (strings.Contains(r.Reason, "because") || strings.Contains(r.Phrase, "auth")) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected causal relation from 'because', got %v", out)
	}
}

func TestExtract_led_to(t *testing.T) {
	content := "The change led to a regression in the API."
	out := Extract(content)
	if len(out) == 0 {
		t.Fatal("expected at least one relation")
	}
	var ledTo Relation
	for _, r := range out {
		if strings.Contains(r.Reason, "led to") {
			ledTo = r
			break
		}
	}
	if ledTo.Phrase == "" {
		t.Errorf("expected phrase from 'led to', got %v", out)
	}
}

func TestExtract_dedup(t *testing.T) {
	content := "Because X. Because X again."
	out := Extract(content)
	phrases := make(map[string]bool)
	for _, r := range out {
		key := r.Phrase
		if phrases[key] {
			t.Errorf("duplicate phrase %q", key)
		}
		phrases[key] = true
	}
}

func TestExtract_truncate(t *testing.T) {
	long := "because "
	for i := 0; i < 300; i++ {
		long += "word "
	}
	out := Extract(long)
	for _, r := range out {
		if len(r.Phrase) > 201 {
			t.Errorf("phrase should be truncated to 200, got len %d", len(r.Phrase))
		}
	}
}

func TestExtract_multiple_patterns(t *testing.T) {
	content := "We did Y after fixing X. This was due to the refactor. Since then we added tests."
	out := Extract(content)
	if len(out) < 2 {
		t.Errorf("expected multiple relations, got %d: %v", len(out), out)
	}
}

func Test_truncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"longer than ten", 10, "longer tha"},
		{"  trimmed  ", 5, "trimm"},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestExtract_reason_capture(t *testing.T) {
	content := "The API failed because the database was down."
	out := Extract(content)
	var becauseRel *Relation
	for i := range out {
		if reflect.DeepEqual(out[i].Phrase, "the database was down") || strings.Contains(out[i].Reason, "because") {
			becauseRel = &out[i]
			break
		}
	}
	if becauseRel == nil {
		t.Logf("Extract output: %v", out)
		t.Fatal("expected one 'because' relation")
	}
	if becauseRel.Phrase == "" {
		t.Errorf("Phrase should be non-empty, got %q", becauseRel.Phrase)
	}
}
