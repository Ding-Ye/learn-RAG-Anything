// Chapter s04 — embedder.
//
// In a real RAG system the embedder calls a hosted model (OpenAI,
// Cohere, Voyage, etc.) to map text to a dense vector. The interface
// is what matters for the rest of the pipeline; the *model* is
// pluggable. This chapter:
//
//   - defines an Embedder interface,
//   - provides FakeEmbedder, a deterministic hash-based vector
//     generator with no dependencies (so go test runs anywhere),
//   - shows how query embeddings and chunk embeddings share the same
//     interface (this matters for retrieval in s05/s06).
//
// Run: go run ./agents/s04-embedder
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// Embedder maps text to a fixed-dimensional vector. The interface is
// intentionally tiny: this is the contract every later chapter relies
// on. A production implementation would batch, handle retries, and
// negotiate a model id; those concerns are out of scope here.
type Embedder interface {
	Dim() int
	Embed(text string) rag.Embedding
}

// FakeEmbedder is a deterministic, dependency-free embedder. It splits
// the input into whitespace tokens, hashes each token, and accumulates
// the hashed pieces into a fixed-dimensional vector. The result is
// stable for a given input and reasonably "smooth" — strings that share
// tokens land near each other in cosine space, which is enough for the
// curriculum's retrieval demo.
type FakeEmbedder struct {
	dim int
}

// NewFakeEmbedder constructs a FakeEmbedder with the given dimension.
// 64 is a good default — large enough to see structure, small enough
// to print.
func NewFakeEmbedder(dim int) *FakeEmbedder {
	if dim <= 0 {
		dim = 64
	}
	return &FakeEmbedder{dim: dim}
}

// Dim reports the vector dimension.
func (f *FakeEmbedder) Dim() int { return f.dim }

// Embed produces the deterministic vector for text.
func (f *FakeEmbedder) Embed(text string) rag.Embedding {
	vec := make(rag.Embedding, f.dim)
	tokens := tokenize(text)
	for _, tok := range tokens {
		sum := sha256.Sum256([]byte(tok))
		// Walk the 32-byte hash in 4-byte windows; map each window to
		// one position in the vector and a small +/- contribution.
		for i := 0; i+4 <= len(sum); i += 4 {
			u := binary.BigEndian.Uint32(sum[i : i+4])
			pos := int(u) % f.dim
			if pos < 0 {
				pos += f.dim
			}
			// Sign bit of the next byte gives us a +/- contribution
			// without bringing in extra randomness.
			sign := float32(1)
			if (u>>31)&1 == 1 {
				sign = -1
			}
			vec[pos] += sign / float32(len(tokens))
		}
	}
	return l2Normalize(vec)
}

// tokenize lowercases the input and splits on whitespace and common
// punctuation. Cheap on purpose.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	fields := strings.FieldsFunc(text, func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', ',', '.', ';', ':', '!', '?', '(', ')', '[', ']', '{', '}', '"', '\'':
			return true
		}
		return false
	})
	if len(fields) == 0 {
		return []string{"__empty__"}
	}
	return fields
}

// l2Normalize returns a unit-length vector. Cosine similarity reduces
// to a dot product once both vectors are normalized — the next chapter
// relies on this.
func l2Normalize(v rag.Embedding) rag.Embedding {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return v
	}
	inv := float32(1.0 / math.Sqrt(sum))
	for i := range v {
		v[i] *= inv
	}
	return v
}

// CosineSimilarity is provided here for the demo; s05 will use it for
// retrieval. Inputs are assumed unit-length.
func CosineSimilarity(a, b rag.Embedding) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}

func main() {
	e := NewFakeEmbedder(64)
	pairs := [][2]string{
		{"retrieval augmented generation", "RAG combines retrieval with generation"},
		{"retrieval augmented generation", "the cat sat on the mat"},
		{"hello world", "hello world"},
	}
	fmt.Printf("Embedder dim=%d\n", e.Dim())
	for _, p := range pairs {
		va := e.Embed(p[0])
		vb := e.Embed(p[1])
		fmt.Printf("  sim=%+.3f  %q  vs  %q\n",
			CosineSimilarity(va, vb), p[0], p[1])
	}
}
