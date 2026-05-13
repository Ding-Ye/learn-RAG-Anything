package main

import (
	"math"
	"testing"
)

func TestFakeEmbedder_Deterministic(t *testing.T) {
	e := NewFakeEmbedder(32)
	a := e.Embed("hello world")
	b := e.Embed("hello world")
	if len(a) != len(b) || len(a) != 32 {
		t.Fatalf("vectors differ in dim: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("vec[%d]: %v vs %v — embedder not deterministic", i, a[i], b[i])
		}
	}
}

func TestFakeEmbedder_UnitLength(t *testing.T) {
	e := NewFakeEmbedder(16)
	v := e.Embed("retrieval augmented generation")
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if math.Abs(math.Sqrt(sum)-1) > 1e-5 {
		t.Errorf("vector not unit length: |v|=%.5f", math.Sqrt(sum))
	}
}

func TestCosineSimilarity_RangeAndIdentity(t *testing.T) {
	e := NewFakeEmbedder(64)
	v := e.Embed("alpha beta gamma")
	if got := CosineSimilarity(v, v); got < 0.999 {
		t.Errorf("self-similarity should be ~1, got %v", got)
	}
}

func TestCosineSimilarity_SharedTokensRankHigher(t *testing.T) {
	e := NewFakeEmbedder(128)
	q := e.Embed("retrieval augmented generation pipeline")
	hits := []struct {
		text string
		want float32
	}{
		{"retrieval augmented generation", 0.4},
		{"vector retrieval pipeline", 0.2},
		{"the cat sat on the mat", -1.0}, // wildcard floor
	}
	scores := make([]float32, len(hits))
	for i, h := range hits {
		scores[i] = CosineSimilarity(q, e.Embed(h.text))
	}
	// The "matching" pair must score higher than the unrelated pair.
	if scores[0] <= scores[2] {
		t.Errorf("matching pair %.3f did not exceed unrelated pair %.3f",
			scores[0], scores[2])
	}
}

func TestTokenize_Empty(t *testing.T) {
	got := tokenize("   ")
	if len(got) != 1 || got[0] != "__empty__" {
		t.Errorf("empty input expected [__empty__], got %v", got)
	}
}
