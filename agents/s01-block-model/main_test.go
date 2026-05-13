package main

import (
	"strings"
	"testing"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

func TestBuildSampleDoc_ShapeAndOrder(t *testing.T) {
	doc := buildSampleDoc()
	if doc.ID == "" || doc.Source == "" {
		t.Fatalf("expected non-empty doc ID and source, got %#v", doc)
	}
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}
	for i, b := range doc.Blocks {
		if b.Order != i {
			t.Errorf("block %d: Order=%d, want %d", i, b.Order, i)
		}
	}
}

func TestSummarize_TaggedByKind(t *testing.T) {
	doc := buildSampleDoc()
	lines := summarize(doc)
	if len(lines) != len(doc.Blocks) {
		t.Fatalf("summarize produced %d lines, want %d", len(lines), len(doc.Blocks))
	}
	tests := []struct {
		idx  int
		kind rag.BlockKind
	}{
		{0, rag.BlockHeading},
		{1, rag.BlockText},
		{2, rag.BlockCode},
	}
	for _, tc := range tests {
		needle := "[" + string(tc.kind) + "]"
		if !strings.Contains(lines[tc.idx], needle) {
			t.Errorf("line %d %q missing %q", tc.idx, lines[tc.idx], needle)
		}
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"short", 10, "short"},
		{"exactlyten", 10, "exactlyten"},
		{"abcdefghijk", 5, "abcd…"},
	}
	for _, tc := range cases {
		got := truncate(tc.in, tc.n)
		if got != tc.want {
			t.Errorf("truncate(%q,%d)=%q, want %q", tc.in, tc.n, got, tc.want)
		}
	}
}
