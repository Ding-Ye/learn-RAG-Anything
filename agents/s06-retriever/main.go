// Chapter s06 — retriever.
//
// A Retriever turns a *natural-language* question into a ranked list
// of chunks. It is the glue between the user's text and the vector
// store. The actual work is small:
//
//   1. embed the query with the same Embedder used to index chunks,
//   2. delegate to the VectorStore for top-k similarity search,
//   3. return the RetrievedChunks with their scores.
//
// What makes this interesting in real systems is *what comes around
// the search*: filtering by metadata, rerankers, query rewriting,
// hybrid retrieval (BM25 + vector). We keep the contract tiny and
// expose a single optional knob: a metadata filter.
//
// Run: go run ./agents/s06-retriever
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// Embedder mirrors the interface from s04. Re-stated here so this
// chapter is self-contained.
type Embedder interface {
	Dim() int
	Embed(text string) rag.Embedding
}

// VectorStore mirrors the interface from s05.
type VectorStore interface {
	Add(rec rag.VectorRecord)
	Search(query rag.Embedding, k int) []rag.RetrievedChunk
	Len() int
}

// Retriever turns text into top-k RetrievedChunks. Behavior is
// deterministic once the Embedder is fixed.
type Retriever struct {
	Embedder Embedder
	Store    VectorStore
}

// RetrieveOptions tunes a single retrieval call.
type RetrieveOptions struct {
	K            int                   // how many results to return; defaults to 5
	SectionFilter string               // optional: only chunks where Section == this value
	MinScore     float32               // discard hits below this similarity floor
	MetaFilter   map[string]string     // optional: chunk.Meta[k]==v must hold for all entries
}

// Retrieve embeds the question, searches the store, and applies the
// optional filters.
func (r *Retriever) Retrieve(question string, opts RetrieveOptions) []rag.RetrievedChunk {
	if opts.K <= 0 {
		opts.K = 5
	}
	// Over-fetch a bit so filters can prune without starving the result.
	overFetch := opts.K * 3
	if overFetch < 10 {
		overFetch = 10
	}
	q := r.Embedder.Embed(question)
	raw := r.Store.Search(q, overFetch)

	out := make([]rag.RetrievedChunk, 0, opts.K)
	for _, hit := range raw {
		if opts.MinScore > 0 && hit.Score < opts.MinScore {
			continue
		}
		if opts.SectionFilter != "" && hit.Chunk.Section != opts.SectionFilter {
			continue
		}
		if !metaMatch(hit.Chunk.Meta, opts.MetaFilter) {
			continue
		}
		out = append(out, hit)
		if len(out) == opts.K {
			break
		}
	}
	// Defensive re-sort: filters preserve order but we re-sort to keep
	// the invariant explicit.
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

func metaMatch(have, want map[string]string) bool {
	for k, v := range want {
		if have[k] != v {
			return false
		}
	}
	return true
}

// --- embedder + store (re-declared so chapter is self-contained) -----

type fakeEmbedder struct{ dim int }

func (f *fakeEmbedder) Dim() int { return f.dim }
func (f *fakeEmbedder) Embed(text string) rag.Embedding {
	vec := make(rag.Embedding, f.dim)
	toks := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	if len(toks) == 0 {
		toks = []string{"__empty__"}
	}
	for _, t := range toks {
		sum := sha256.Sum256([]byte(t))
		for i := 0; i+4 <= len(sum); i += 4 {
			u := binary.BigEndian.Uint32(sum[i : i+4])
			pos := int(u) % f.dim
			if pos < 0 {
				pos += f.dim
			}
			sign := float32(1)
			if (u>>31)&1 == 1 {
				sign = -1
			}
			vec[pos] += sign / float32(len(toks))
		}
	}
	var s float64
	for _, x := range vec {
		s += float64(x) * float64(x)
	}
	if s == 0 {
		return vec
	}
	inv := float32(1.0 / math.Sqrt(s))
	for i := range vec {
		vec[i] *= inv
	}
	return vec
}

type memStore struct{ recs []rag.VectorRecord }

func (s *memStore) Add(r rag.VectorRecord) { s.recs = append(s.recs, r) }
func (s *memStore) Len() int               { return len(s.recs) }
func (s *memStore) Search(q rag.Embedding, k int) []rag.RetrievedChunk {
	scored := make([]rag.RetrievedChunk, 0, len(s.recs))
	for _, r := range s.recs {
		var d float32
		for i := range q {
			d += q[i] * r.Embedding[i]
		}
		scored = append(scored, rag.RetrievedChunk{Chunk: r.Chunk, Score: d})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	if k > len(scored) {
		k = len(scored)
	}
	return scored[:k]
}

func main() {
	emb := &fakeEmbedder{dim: 64}
	store := &memStore{}

	corpus := []rag.Chunk{
		{ID: "c1", Text: "Retrieval finds top-k similar chunks", Section: "Pipeline"},
		{ID: "c2", Text: "Chunks are embedded into dense vectors", Section: "Pipeline"},
		{ID: "c3", Text: "Markdown headings become heading blocks", Section: "Parsing"},
		{ID: "c4", Text: "Page numbers travel with each block", Section: "Parsing"},
		{ID: "c5", Text: "RAG combines retrieval with generation", Section: "Pipeline"},
	}
	for _, c := range corpus {
		store.Add(rag.VectorRecord{ChunkID: c.ID, Embedding: emb.Embed(c.Text), Chunk: c})
	}

	r := &Retriever{Embedder: emb, Store: store}

	for _, q := range []string{
		"How does retrieval work?",
		"How are pages tracked?",
	} {
		hits := r.Retrieve(q, RetrieveOptions{K: 3})
		fmt.Printf("\nQ: %q\n", q)
		for i, h := range hits {
			fmt.Printf("  #%d score=%+.3f section=%-10q text=%q\n",
				i+1, h.Score, h.Chunk.Section, h.Chunk.Text)
		}
	}

	// Filtered retrieval: only "Parsing" section.
	hits := r.Retrieve("blocks", RetrieveOptions{K: 5, SectionFilter: "Parsing"})
	fmt.Printf("\nQ: %q (section=Parsing)\n", "blocks")
	for i, h := range hits {
		fmt.Printf("  #%d score=%+.3f text=%q\n", i+1, h.Score, h.Chunk.Text)
	}
}
