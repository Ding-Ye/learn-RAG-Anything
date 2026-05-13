package main

import (
	"strings"
	"testing"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

const tinyDoc = `# Alpha

Alpha section talks about retrieval and ranking.

## Beta

Beta section explains embeddings and vector spaces.

## Gamma

Gamma section talks about cats and mats.
`

func TestPipeline_IngestProducesChunks(t *testing.T) {
	p := NewPipeline()
	n := p.Ingest("d", "d.md", tinyDoc)
	if n == 0 {
		t.Fatalf("Ingest should produce at least one chunk, got 0")
	}
	if p.Store.Len() != n {
		t.Errorf("Store.Len=%d, want %d", p.Store.Len(), n)
	}
}

func TestPipeline_AskReturnsAnswerAndHits(t *testing.T) {
	p := NewPipeline()
	p.Ingest("d", "d.md", tinyDoc)

	ans, hits, prompt := p.Ask("Tell me about embeddings", 3)
	if len(hits) == 0 {
		t.Fatalf("expected hits, got 0")
	}
	if !strings.Contains(prompt, "CONTEXT:") || !strings.Contains(prompt, "QUESTION:") {
		t.Errorf("prompt malformed:\n%s", prompt)
	}
	if !strings.Contains(ans, "(mock") {
		t.Errorf("expected mock answer marker, got %q", ans)
	}
}

func TestPipeline_BetaSectionMatchesEmbeddingsQuery(t *testing.T) {
	p := NewPipeline()
	p.Ingest("d", "d.md", tinyDoc)
	_, hits, _ := p.Ask("embeddings and vectors", 5)
	if len(hits) == 0 {
		t.Fatalf("no hits returned")
	}
	top := hits[0]
	if !(strings.Contains(strings.ToLower(top.Chunk.Text), "embed") ||
		strings.EqualFold(top.Chunk.Section, "Beta")) {
		t.Errorf("top hit unrelated to query: %+v", top.Chunk)
	}
}

func TestRenderPrompt_OrderAndCitations(t *testing.T) {
	hits := []rag.RetrievedChunk{
		{Chunk: rag.Chunk{Text: "MARK_HIGH"}, Score: 0.9},
		{Chunk: rag.Chunk{Text: "MARK_MID"}, Score: 0.6},
		{Chunk: rag.Chunk{Text: "MARK_LOW"}, Score: 0.3},
	}
	prompt := renderPrompt("q?", hits)
	if !strings.Contains(prompt, "[c1] MARK_HIGH") {
		t.Errorf("first citation missing or wrong:\n%s", prompt)
	}
	idxHigh := strings.Index(prompt, "MARK_HIGH")
	idxMid := strings.Index(prompt, "MARK_MID")
	idxLow := strings.Index(prompt, "MARK_LOW")
	if !(idxHigh < idxMid && idxMid < idxLow) {
		t.Errorf("ordering wrong in prompt: high=%d mid=%d low=%d", idxHigh, idxMid, idxLow)
	}
}

func TestParseAndChunk_HeadingsArePreserved(t *testing.T) {
	blocks := parseMarkdown(tinyDoc)
	var headings int
	for _, b := range blocks {
		if b.Kind == rag.BlockHeading {
			headings++
		}
	}
	if headings != 3 {
		t.Errorf("expected 3 headings (Alpha/Beta/Gamma), got %d", headings)
	}
}

func TestMockLLM_NoContext(t *testing.T) {
	got := MockLLM{}.Complete("no headers here")
	if !strings.Contains(got, "does not say") {
		t.Errorf("MockLLM should fall back when no CONTEXT block, got %q", got)
	}
}
