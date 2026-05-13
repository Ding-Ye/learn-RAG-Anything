package main

import (
	"strings"
	"testing"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

func makeTextDoc(blocks ...rag.Block) rag.Document {
	for i := range blocks {
		blocks[i].Order = i
		if blocks[i].Page == 0 {
			blocks[i].Page = 1
		}
	}
	return rag.Document{ID: "t", Source: "t.md", Blocks: blocks}
}

func TestChunk_SplitsAtTarget(t *testing.T) {
	long := strings.Repeat("abcdefghij ", 30) // ~330 chars
	doc := makeTextDoc(rag.Block{Kind: rag.BlockText, Text: long, Section: "S"})
	out := Chunk(doc, ChunkConfig{TargetChars: 80, OverlapChars: 10})
	if len(out) < 2 {
		t.Fatalf("expected multiple chunks for 330-char input, got %d", len(out))
	}
	for i, c := range out {
		if len(c.Text) > 200 {
			t.Errorf("chunk %d too big: %d chars", i, len(c.Text))
		}
		if c.DocID != "t" {
			t.Errorf("chunk %d missing DocID", i)
		}
	}
}

func TestChunk_HeadingIsItsOwnChunk(t *testing.T) {
	doc := makeTextDoc(
		rag.Block{Kind: rag.BlockHeading, Text: "Title", Section: "Title"},
		rag.Block{Kind: rag.BlockText, Text: "body text under title.", Section: "Title"},
	)
	out := Chunk(doc, ChunkConfig{TargetChars: 200, OverlapChars: 10})
	if len(out) != 2 {
		t.Fatalf("want 2 chunks, got %d", len(out))
	}
	if out[0].Text != "Title" {
		t.Errorf("first chunk should be heading 'Title', got %q", out[0].Text)
	}
	if !strings.Contains(out[1].Text, "body text under title") {
		t.Errorf("second chunk lost body: %q", out[1].Text)
	}
}

func TestChunk_CodeStaysWhole(t *testing.T) {
	code := strings.Repeat("func() { return }\n", 20)
	doc := makeTextDoc(rag.Block{Kind: rag.BlockCode, Text: code, Lang: "go"})
	out := Chunk(doc, ChunkConfig{TargetChars: 50, OverlapChars: 0})
	if len(out) != 1 {
		t.Fatalf("code should not be split: got %d chunks", len(out))
	}
	if !strings.Contains(out[0].Text, "func() { return }") {
		t.Errorf("code body altered:\n got=%q", out[0].Text)
	}
	// Code chunk must not be split: count occurrences of the marker
	// across all chunks; it should equal the source's count.
	want := strings.Count(code, "func() { return }")
	got := strings.Count(out[0].Text, "func() { return }")
	if got != want {
		t.Errorf("code body lost lines: got %d occurrences, want %d", got, want)
	}
}

func TestChunk_OverlapShared(t *testing.T) {
	text := strings.Repeat("xyzpdq ", 60) // ~420 chars
	doc := makeTextDoc(rag.Block{Kind: rag.BlockText, Text: text})
	out := Chunk(doc, ChunkConfig{TargetChars: 60, OverlapChars: 20})
	if len(out) < 2 {
		t.Fatalf("need at least 2 chunks for overlap test, got %d", len(out))
	}
	prevTail := out[0].Text[len(out[0].Text)-20:]
	if !strings.HasPrefix(out[1].Text, prevTail[:min(len(prevTail), 10)]) {
		// Use a partial prefix match — flushing trims whitespace, so
		// the overlap may shift slightly. This is the property we want
		// to assert: meaningful tail-prefix overlap.
		t.Errorf("expected overlap between chunks; tail=%q head=%q",
			prevTail, out[1].Text[:min(len(out[1].Text), 30)])
	}
}

func TestDedupPages(t *testing.T) {
	got := dedupPages([]int{1, 1, 2, 2, 3})
	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("dedupPages[%d]=%d want %d", i, got[i], want[i])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
