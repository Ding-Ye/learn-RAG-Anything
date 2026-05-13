package main

import (
	"testing"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

func seed(t *testing.T, dim int) *MemoryStore {
	t.Helper()
	store := NewMemoryStore()
	for _, c := range []rag.Chunk{
		{ID: "a", Text: "alpha beta gamma"},
		{ID: "b", Text: "delta epsilon"},
		{ID: "c", Text: "alpha beta"},
		{ID: "d", Text: "completely unrelated mango banana"},
	} {
		store.Add(rag.VectorRecord{
			ChunkID:   c.ID,
			Embedding: tinyEmbedder(c.Text, dim),
			Chunk:     c,
		})
	}
	return store
}

func TestMemoryStore_AddLen(t *testing.T) {
	s := NewMemoryStore()
	if s.Len() != 0 {
		t.Errorf("new store should have Len 0, got %d", s.Len())
	}
	s.Add(rag.VectorRecord{ChunkID: "x"})
	if s.Len() != 1 {
		t.Errorf("after Add Len should be 1, got %d", s.Len())
	}
}

func TestMemoryStore_SearchRanksMatching(t *testing.T) {
	store := seed(t, 64)
	q := tinyEmbedder("alpha beta gamma", 64)
	hits := store.Search(q, 4)
	if len(hits) != 4 {
		t.Fatalf("expected 4 hits, got %d", len(hits))
	}
	// 'a' is identical to the query and must rank first.
	if hits[0].Chunk.ID != "a" {
		t.Errorf("top hit should be 'a' (identical), got %q (score=%.3f)", hits[0].Chunk.ID, hits[0].Score)
	}
	// The unrelated chunk must rank below the partial match.
	last := hits[len(hits)-1]
	if last.Chunk.ID != "d" {
		t.Errorf("bottom hit should be 'd' (unrelated), got %q", last.Chunk.ID)
	}
}

func TestMemoryStore_SearchK(t *testing.T) {
	store := seed(t, 32)
	q := tinyEmbedder("alpha", 32)
	for _, k := range []int{1, 2, 4, 10} {
		hits := store.Search(q, k)
		want := k
		if want > store.Len() {
			want = store.Len()
		}
		if len(hits) != want {
			t.Errorf("k=%d: got %d hits, want %d", k, len(hits), want)
		}
	}
}

func TestMemoryStore_SearchEmpty(t *testing.T) {
	s := NewMemoryStore()
	hits := s.Search(tinyEmbedder("foo", 16), 5)
	if len(hits) != 0 {
		t.Errorf("empty store should return no hits, got %d", len(hits))
	}
}

func TestDotMismatchedDim(t *testing.T) {
	if got := dot(rag.Embedding{1, 0}, rag.Embedding{1, 0, 0}); got != 0 {
		t.Errorf("dot on mismatched dims should be 0, got %v", got)
	}
}
