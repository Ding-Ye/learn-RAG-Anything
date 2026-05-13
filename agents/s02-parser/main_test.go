package main

import (
	"strings"
	"testing"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

func TestParseMarkdown_Tables(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []rag.BlockKind
	}{
		{
			name: "single heading and paragraph",
			src:  "# Title\n\nHello world.\n",
			want: []rag.BlockKind{rag.BlockHeading, rag.BlockText},
		},
		{
			name: "code fence with language",
			src:  "```go\nfmt.Println(\"hi\")\n```\n",
			want: []rag.BlockKind{rag.BlockCode},
		},
		{
			name: "page break increments page",
			src:  "p1.\n\n---\n\np2.\n",
			want: []rag.BlockKind{rag.BlockText, rag.BlockText},
		},
		{
			name: "empty input produces no blocks",
			src:  "\n\n\n",
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseMarkdown(tc.src)
			if len(got) != len(tc.want) {
				t.Fatalf("len(blocks)=%d, want %d (%+v)", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i].Kind != tc.want[i] {
					t.Errorf("block %d kind=%s, want %s", i, got[i].Kind, tc.want[i])
				}
			}
		})
	}
}

func TestParseMarkdown_PageBreaks(t *testing.T) {
	src := "first.\n\n---\n\nsecond.\n"
	got := ParseMarkdown(src)
	if len(got) != 2 {
		t.Fatalf("want 2 blocks, got %d", len(got))
	}
	if got[0].Page != 1 || got[1].Page != 2 {
		t.Errorf("page numbers: got[0]=%d got[1]=%d, want 1 and 2", got[0].Page, got[1].Page)
	}
}

func TestParseMarkdown_SectionTracking(t *testing.T) {
	src := "# Alpha\n\ntext under alpha.\n\n## Beta\n\ntext under beta.\n"
	got := ParseMarkdown(src)
	// expected: heading, text, heading, text
	if len(got) != 4 {
		t.Fatalf("want 4 blocks, got %d", len(got))
	}
	if got[1].Section != "Alpha" {
		t.Errorf("got[1].Section=%q, want Alpha", got[1].Section)
	}
	if got[3].Section != "Beta" {
		t.Errorf("got[3].Section=%q, want Beta", got[3].Section)
	}
}

func TestParseMarkdown_SampleDoc(t *testing.T) {
	got := ParseMarkdown(sampleDoc)
	var foundCode bool
	for _, b := range got {
		if b.Kind == rag.BlockCode && strings.Contains(b.Text, "store.Search") {
			foundCode = true
		}
	}
	if !foundCode {
		t.Errorf("sample doc should produce a code block containing store.Search; got=%+v", got)
	}
}
