// Chapter s05 — vector store.
//
// An in-memory vector store: an unordered slice of (chunkID,
// embedding, chunk) records, with a top-k search by cosine
// similarity. Real systems put this in pgvector / FAISS / Qdrant /
// pinecone; the storage layer is interchangeable, but the *interface*
// is universal: Add and Search.
//
// Run: go run ./agents/s05-vector-store
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// VectorStore is the interface every later chapter uses. We could swap
// the in-memory impl for a network-backed one without touching anything
// else.
type VectorStore interface {
	Add(rec rag.VectorRecord)
	Search(query rag.Embedding, k int) []rag.RetrievedChunk
	Len() int
}

// MemoryStore is a tiny in-memory VectorStore. Linear scan; fine up to
// the low thousands of records, which is plenty for a teaching repo.
type MemoryStore struct {
	records []rag.VectorRecord
}

// NewMemoryStore returns an empty store.
func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

// Add appends a record. We do not deduplicate; the caller is expected
// to use stable Chunk.ID values.
func (s *MemoryStore) Add(rec rag.VectorRecord) {
	s.records = append(s.records, rec)
}

// Len reports the number of records in the store.
func (s *MemoryStore) Len() int { return len(s.records) }

// Search returns the top-k records by cosine similarity to query.
// Embeddings are assumed unit-length (see s04 FakeEmbedder), so cosine
// reduces to a dot product.
func (s *MemoryStore) Search(query rag.Embedding, k int) []rag.RetrievedChunk {
	if k <= 0 || len(s.records) == 0 {
		return nil
	}
	scored := make([]rag.RetrievedChunk, 0, len(s.records))
	for _, r := range s.records {
		score := dot(query, r.Embedding)
		scored = append(scored, rag.RetrievedChunk{
			Chunk: r.Chunk,
			Score: score,
		})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	if k > len(scored) {
		k = len(scored)
	}
	return scored[:k]
}

func dot(a, b rag.Embedding) float32 {
	if len(a) != len(b) {
		return 0
	}
	var s float32
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

// --- demo glue --------------------------------------------------------

// tinyEmbedder mirrors the s04 FakeEmbedder so this chapter is
// self-contained. Real consumers will import the s04 implementation;
// we duplicate here on purpose so each chapter can be read and run
// stand-alone.
func tinyEmbedder(text string, dim int) rag.Embedding {
	vec := make(rag.Embedding, dim)
	for _, tok := range strings.Fields(strings.ToLower(text)) {
		var h uint32 = 2166136261
		for i := 0; i < len(tok); i++ {
			h ^= uint32(tok[i])
			h *= 16777619
		}
		pos := int(h) % dim
		if pos < 0 {
			pos += dim
		}
		vec[pos] += 1
	}
	// L2 normalize.
	var sum float64
	for _, x := range vec {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return vec
	}
	inv := float32(1.0 / sqrt(sum))
	for i := range vec {
		vec[i] *= inv
	}
	return vec
}

func sqrt(x float64) float64 {
	// Newton's method, two iterations — keeps this chapter dependency-free.
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 8; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func main() {
	const dim = 32
	store := NewMemoryStore()

	corpus := []rag.Chunk{
		{ID: "a", Text: "RAG combines retrieval with text generation"},
		{ID: "b", Text: "Cosine similarity is dot product on unit vectors"},
		{ID: "c", Text: "The cat sat on the mat"},
		{ID: "d", Text: "Embeddings map text to dense vectors"},
		{ID: "e", Text: "Retrieval picks top-k chunks by similarity"},
	}
	for _, c := range corpus {
		store.Add(rag.VectorRecord{
			ChunkID:   c.ID,
			Embedding: tinyEmbedder(c.Text, dim),
			Chunk:     c,
		})
	}
	fmt.Printf("indexed %d chunks (dim=%d)\n", store.Len(), dim)

	query := "retrieval and generation"
	qvec := tinyEmbedder(query, dim)
	hits := store.Search(qvec, 3)
	fmt.Printf("\nquery: %q\n", query)
	for i, h := range hits {
		fmt.Printf("  #%d score=%+.3f id=%s text=%q\n", i+1, h.Score, h.Chunk.ID, h.Chunk.Text)
	}
}
