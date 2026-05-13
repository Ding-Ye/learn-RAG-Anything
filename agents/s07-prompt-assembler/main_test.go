package main

import (
	"strings"
	"testing"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

func sampleHits() []rag.RetrievedChunk {
	return []rag.RetrievedChunk{
		{Chunk: rag.Chunk{ID: "a", Text: "first chunk about RAG", Section: "Intro"}, Score: 0.9},
		{Chunk: rag.Chunk{ID: "b", Text: "second chunk about chunks", Section: "Pipeline"}, Score: 0.6},
		{Chunk: rag.Chunk{ID: "c", Text: "noisy off-topic chunk", Section: "Other"}, Score: 0.05},
	}
}

func TestAssemble_MinScoreDrops(t *testing.T) {
	p := AssemblePrompt("q?", sampleHits(), AssembleOptions{MinScore: 0.2})
	if len(p.UsedChunks) != 2 {
		t.Errorf("MinScore=0.2 should keep 2 chunks, got %d", len(p.UsedChunks))
	}
}

func TestAssemble_CitationMarkers(t *testing.T) {
	p := AssemblePrompt("q?", sampleHits(), AssembleOptions{IncludeIDs: true})
	if !strings.Contains(p.Text, "[c1]") || !strings.Contains(p.Text, "[c2]") {
		t.Errorf("citation markers missing:\n%s", p.Text)
	}
}

func TestAssemble_MaxCharsTruncates(t *testing.T) {
	hits := []rag.RetrievedChunk{
		{Chunk: rag.Chunk{ID: "a", Text: strings.Repeat("x", 80)}, Score: 0.9},
		{Chunk: rag.Chunk{ID: "b", Text: strings.Repeat("y", 80)}, Score: 0.5},
		{Chunk: rag.Chunk{ID: "c", Text: strings.Repeat("z", 80)}, Score: 0.3},
	}
	p := AssemblePrompt("q?", hits, AssembleOptions{MaxChars: 120})
	if p.DroppedCount == 0 {
		t.Errorf("MaxChars=120 with 3x80 chunks should drop something, got dropped=0")
	}
	if len(p.UsedChunks) >= len(hits) {
		t.Errorf("truncation should reduce chunks; have %d", len(p.UsedChunks))
	}
}

func TestAssemble_LangZH(t *testing.T) {
	p := AssemblePrompt("RAG 是什么？", sampleHits(), AssembleOptions{Lang: LangZH})
	if !strings.Contains(p.Text, "【参考资料】") {
		t.Errorf("ZH prompt missing reference header:\n%s", p.Text)
	}
}

func TestAssemble_LangEN(t *testing.T) {
	p := AssemblePrompt("what is RAG?", sampleHits(), AssembleOptions{})
	if !strings.Contains(p.Text, "CONTEXT:") || !strings.Contains(p.Text, "QUESTION:") {
		t.Errorf("EN prompt should have CONTEXT: and QUESTION:\n%s", p.Text)
	}
}

func TestAssemble_PreservesScoreOrder(t *testing.T) {
	hits := []rag.RetrievedChunk{
		{Chunk: rag.Chunk{ID: "low", Text: "ZZZ-low-marker"}, Score: 0.3},
		{Chunk: rag.Chunk{ID: "high", Text: "ZZZ-high-marker"}, Score: 0.9},
		{Chunk: rag.Chunk{ID: "mid", Text: "ZZZ-mid-marker"}, Score: 0.6},
	}
	p := AssemblePrompt("q?", hits, AssembleOptions{})
	idxHigh := strings.Index(p.Text, "ZZZ-high-marker")
	idxMid := strings.Index(p.Text, "ZZZ-mid-marker")
	idxLow := strings.Index(p.Text, "ZZZ-low-marker")
	if !(idxHigh < idxMid && idxMid < idxLow) {
		t.Errorf("score order not preserved in prompt (idx high=%d mid=%d low=%d):\n%s",
			idxHigh, idxMid, idxLow, p.Text)
	}
}
