package main

import (
	"testing"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

func seed(t *testing.T) *Retriever {
	t.Helper()
	emb := &fakeEmbedder{dim: 64}
	store := &memStore{}
	for _, c := range []rag.Chunk{
		{ID: "x1", Text: "alpha beta gamma", Section: "A", Meta: map[string]string{"kind": "text"}},
		{ID: "x2", Text: "alpha beta delta", Section: "B", Meta: map[string]string{"kind": "text"}},
		{ID: "x3", Text: "epsilon zeta", Section: "A", Meta: map[string]string{"kind": "code"}},
		{ID: "x4", Text: "mango banana orange", Section: "C", Meta: map[string]string{"kind": "text"}},
	} {
		store.Add(rag.VectorRecord{ChunkID: c.ID, Embedding: emb.Embed(c.Text), Chunk: c})
	}
	return &Retriever{Embedder: emb, Store: store}
}

func TestRetrieve_KCap(t *testing.T) {
	r := seed(t)
	hits := r.Retrieve("alpha beta", RetrieveOptions{K: 2})
	if len(hits) != 2 {
		t.Errorf("K=2 should yield 2 results, got %d", len(hits))
	}
}

func TestRetrieve_SectionFilter(t *testing.T) {
	r := seed(t)
	hits := r.Retrieve("alpha beta", RetrieveOptions{K: 5, SectionFilter: "A"})
	for _, h := range hits {
		if h.Chunk.Section != "A" {
			t.Errorf("filter violation: section=%q", h.Chunk.Section)
		}
	}
	if len(hits) == 0 {
		t.Errorf("section=A should match at least one chunk")
	}
}

func TestRetrieve_MetaFilter(t *testing.T) {
	r := seed(t)
	hits := r.Retrieve("alpha", RetrieveOptions{
		K: 5, MetaFilter: map[string]string{"kind": "code"},
	})
	for _, h := range hits {
		if h.Chunk.Meta["kind"] != "code" {
			t.Errorf("meta filter violation: meta=%v", h.Chunk.Meta)
		}
	}
}

func TestRetrieve_MinScore(t *testing.T) {
	r := seed(t)
	hits := r.Retrieve("mango", RetrieveOptions{K: 5, MinScore: 0.999})
	for _, h := range hits {
		if h.Score < 0.999 {
			t.Errorf("min-score violation: %v", h.Score)
		}
	}
}

func TestRetrieve_DefaultsK(t *testing.T) {
	r := seed(t)
	hits := r.Retrieve("alpha", RetrieveOptions{})
	if len(hits) != 4 {
		// We only have 4 chunks total; default K is 5 → capped at corpus size.
		t.Errorf("default K should clamp to corpus size 4, got %d", len(hits))
	}
}

func TestMetaMatch(t *testing.T) {
	if !metaMatch(map[string]string{"a": "1"}, nil) {
		t.Errorf("nil want should always match")
	}
	if !metaMatch(map[string]string{"a": "1", "b": "2"}, map[string]string{"a": "1"}) {
		t.Errorf("subset match should pass")
	}
	if metaMatch(map[string]string{"a": "1"}, map[string]string{"a": "2"}) {
		t.Errorf("differing value should fail")
	}
	if metaMatch(map[string]string{}, map[string]string{"a": "1"}) {
		t.Errorf("missing key should fail")
	}
}
