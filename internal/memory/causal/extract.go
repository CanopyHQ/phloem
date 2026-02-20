// Package causal provides causal relation extraction from memory content.
// Used by the memory store to build causal edges between memories (Stage 1).
package causal

import (
	"regexp"
	"strings"
)

// Relation represents a causal relation extracted from text (e.g. "because X", "led to Y").
// Phrase is used to search for a related memory; Reason is the full causal snippet.
type Relation struct {
	Phrase string // substring to use for semantic search when linking to another memory
	Reason string // full causal phrase (e.g. "because we fixed the auth bug")
}

// causalPatterns match common causal language; capture group is the cause/effect phrase
var causalPatterns = []struct {
	re   *regexp.Regexp
	name string
}{
	{regexp.MustCompile(`(?i)\bbecause\s+(.+?)(?:\.|,|;|\s+and\s+|\s+so\s+|$)`), "because"},
	{regexp.MustCompile(`(?i)\bso\s+that\s+(.+?)(?:\.|,|;|$)`), "so_that"},
	{regexp.MustCompile(`(?i)\bcaused\s+by\s+(.+?)(?:\.|,|;|$)`), "caused_by"},
	{regexp.MustCompile(`(?i)\bled\s+to\s+(.+?)(?:\.|,|;|$)`), "led_to"},
	{regexp.MustCompile(`(?i)\bafter\s+(.+?)(?:\s+we\s+|\s+,|\.|$)`), "after"},
	{regexp.MustCompile(`(?i)\bdue\s+to\s+(.+?)(?:\.|,|;|$)`), "due_to"},
	{regexp.MustCompile(`(?i)\bsince\s+(.+?)(?:\.|,|;|$)`), "since"},
	{regexp.MustCompile(`(?i)\bin\s+order\s+to\s+(.+?)(?:\.|,|;|$)`), "in_order_to"},
}

const maxPhraseLen = 200
const minPhraseLen = 3

// Extract finds causal relations in content using simple pattern matching.
// Returns a deduplicated list of relations; Phrase is trimmed for use as a recall query.
func Extract(content string) []Relation {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	seen := make(map[string]bool)
	var out []Relation
	for _, p := range causalPatterns {
		matches := p.re.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			phrase := strings.TrimSpace(m[1])
			phrase = truncate(phrase, maxPhraseLen)
			if len(phrase) < minPhraseLen {
				continue
			}
			key := strings.ToLower(phrase)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, Relation{Phrase: phrase, Reason: strings.TrimSpace(m[0])})
		}
	}
	return out
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max]
}
