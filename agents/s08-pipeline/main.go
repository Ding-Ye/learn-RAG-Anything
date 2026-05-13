// Chapter s08 — end-to-end pipeline.
//
// Compose everything into a tiny RAG runtime:
//
//   parse → chunk → embed → store → retrieve → assemble → (mock LLM)
//
// The previous chapters live as their own packages-of-one
// (package main) so they can each be run with `go run`. To keep the
// pipeline self-contained without re-importing those mains, we
// re-declare the minimal helpers here. Each helper is a faithful
// shrink of its chapter; the comments point you back.
//
// Run: go run ./agents/s08-pipeline
package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// ---------------------------------------------------------------
// MockLLM — stands in for an OpenAI/Anthropic call.
//
// All it does is print the prompt and return a synthetic answer that
// quotes the first context chunk. This is enough to demonstrate
// end-to-end flow; replacing it with a real LLM is a one-function
// change.
// ---------------------------------------------------------------

type LLM interface {
	Complete(prompt string) string
}

type MockLLM struct{}

func (MockLLM) Complete(prompt string) string {
	// Pull out the first context line as a synthetic "answer".
	idx := strings.Index(prompt, "CONTEXT:\n")
	if idx < 0 {
		return "(mock) The context does not say."
	}
	rest := prompt[idx+len("CONTEXT:\n"):]
	lines := strings.SplitN(rest, "\n", 2)
	first := strings.TrimSpace(lines[0])
	return "(mock answer) Based on the top-ranked chunk, " + first
}

// ---------------------------------------------------------------
// Pipeline — the public surface of this chapter.
// ---------------------------------------------------------------

type Pipeline struct {
	Embedder Embedder
	Store    VectorStore
	LLM      LLM
}

// NewPipeline wires defaults: 64-dim fake embedder, empty memory
// store, mock LLM. Each can be overridden by setting the fields after
// construction.
func NewPipeline() *Pipeline {
	return &Pipeline{
		Embedder: newFakeEmbedder(64),
		Store:    newMemStore(),
		LLM:      MockLLM{},
	}
}

// Ingest runs parse → chunk → embed → store for one markdown blob.
func (p *Pipeline) Ingest(docID, source, markdown string) int {
	blocks := parseMarkdown(markdown)
	doc := rag.Document{ID: docID, Source: source, Blocks: blocks}
	chunks := chunkDoc(doc, chunkConfig{TargetChars: 220, OverlapChars: 30})
	for _, c := range chunks {
		p.Store.Add(rag.VectorRecord{
			ChunkID:   c.ID,
			Embedding: p.Embedder.Embed(c.Text),
			Chunk:     c,
		})
	}
	return len(chunks)
}

// Ask runs embed → retrieve → assemble → LLM for one question.
func (p *Pipeline) Ask(question string, k int) (answer string, hits []rag.RetrievedChunk, prompt string) {
	q := p.Embedder.Embed(question)
	hits = p.Store.Search(q, k)
	prompt = renderPrompt(question, hits)
	answer = p.LLM.Complete(prompt)
	return
}

// ---------------------------------------------------------------
// Compressed re-implementations of s02–s07 (each ≤ 30 lines).
// ---------------------------------------------------------------

// s02 parser — markdown → []Block. Handles headings, paragraphs, fences.
func parseMarkdown(src string) []rag.Block {
	lines := strings.Split(src, "\n")
	var (
		blocks    []rag.Block
		paragraph []string
		section   string
		page      = 1
		order     int
	)
	flush := func() {
		if len(paragraph) == 0 {
			return
		}
		blocks = append(blocks, rag.Block{
			Kind: rag.BlockText, Text: strings.TrimSpace(strings.Join(paragraph, " ")),
			Page: page, Order: order, Section: section,
		})
		order++
		paragraph = paragraph[:0]
	}
	for i := 0; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		switch {
		case t == "---":
			flush()
			page++
		case strings.HasPrefix(t, "```"):
			flush()
			body := []string{}
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				body = append(body, lines[i])
				i++
			}
			blocks = append(blocks, rag.Block{
				Kind: rag.BlockCode, Text: strings.Join(body, "\n"),
				Page: page, Order: order, Section: section,
			})
			order++
		case strings.HasPrefix(t, "#"):
			flush()
			title := strings.TrimSpace(strings.TrimLeft(t, "#"))
			section = title
			blocks = append(blocks, rag.Block{Kind: rag.BlockHeading, Text: title, Page: page, Order: order, Section: section})
			order++
		case t == "":
			flush()
		default:
			paragraph = append(paragraph, t)
		}
	}
	flush()
	return blocks
}

// s03 chunker — []Block → []Chunk with overlap.
type chunkConfig struct{ TargetChars, OverlapChars int }

func chunkDoc(doc rag.Document, cfg chunkConfig) []rag.Chunk {
	var (
		out      []rag.Chunk
		buf      strings.Builder
		section  string
		idx      int
		bufPages []int
	)
	flush := func() {
		text := strings.TrimSpace(buf.String())
		if text == "" {
			return
		}
		out = append(out, rag.Chunk{
			ID:    fmt.Sprintf("%s-%03d-%s", doc.ID, idx, hashHex(doc.ID+"|"+text, 4)),
			Text:  text, DocID: doc.ID, Section: section, Pages: dedupInts(bufPages),
		})
		idx++
		tail := text
		if cfg.OverlapChars > 0 && len(text) > cfg.OverlapChars {
			tail = text[len(text)-cfg.OverlapChars:]
		} else {
			tail = ""
		}
		buf.Reset()
		bufPages = bufPages[:0]
		buf.WriteString(tail)
	}
	for _, b := range doc.Blocks {
		if b.Kind == rag.BlockHeading || b.Kind == rag.BlockCode {
			flush()
			sec := b.Section
			if b.Kind == rag.BlockHeading {
				sec = b.Text
			}
			out = append(out, rag.Chunk{
				ID: fmt.Sprintf("%s-%03d-%s", doc.ID, idx, hashHex(doc.ID+"|"+b.Text, 4)),
				Text: b.Text, DocID: doc.ID, Section: sec, Pages: []int{b.Page},
			})
			idx++
			section = sec
			continue
		}
		section = b.Section
		rem := b.Text
		for len(rem) > 0 {
			room := cfg.TargetChars - buf.Len()
			if room <= 0 {
				flush()
				continue
			}
			take := room
			if take > len(rem) {
				take = len(rem)
			}
			if buf.Len() > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(rem[:take])
			rem = rem[take:]
			bufPages = append(bufPages, b.Page)
			if buf.Len() >= cfg.TargetChars {
				flush()
			}
		}
	}
	flush()
	return out
}

// s04 embedder — deterministic hash-based, unit length.
type Embedder interface {
	Dim() int
	Embed(text string) rag.Embedding
}
type fakeEmbedder struct{ dim int }

func newFakeEmbedder(dim int) *fakeEmbedder { return &fakeEmbedder{dim: dim} }
func (f *fakeEmbedder) Dim() int            { return f.dim }
func (f *fakeEmbedder) Embed(text string) rag.Embedding {
	vec := make(rag.Embedding, f.dim)
	toks := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	if len(toks) == 0 {
		toks = []string{"_e_"}
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

// s05 vector store — linear scan, cosine via dot product.
type VectorStore interface {
	Add(r rag.VectorRecord)
	Search(q rag.Embedding, k int) []rag.RetrievedChunk
	Len() int
}
type memStore struct{ recs []rag.VectorRecord }

func newMemStore() *memStore       { return &memStore{} }
func (s *memStore) Add(r rag.VectorRecord) { s.recs = append(s.recs, r) }
func (s *memStore) Len() int       { return len(s.recs) }
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

// s07 prompt — simple EN template with citation markers.
func renderPrompt(question string, hits []rag.RetrievedChunk) string {
	var b strings.Builder
	b.WriteString("You are a careful assistant. Answer the QUESTION using ONLY the CONTEXT.\n\n")
	b.WriteString("CONTEXT:\n")
	for i, h := range hits {
		fmt.Fprintf(&b, "[c%d] %s\n", i+1, strings.TrimSpace(h.Chunk.Text))
	}
	fmt.Fprintf(&b, "\nQUESTION:\n%s\n\nANSWER:\n", question)
	return b.String()
}

// ---------------------------------------------------------------
// Small utilities used above.
// ---------------------------------------------------------------

func hashHex(s string, nbytes int) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:nbytes])
}

func dedupInts(in []int) []int {
	if len(in) == 0 {
		return nil
	}
	seen := map[int]struct{}{}
	out := make([]int, 0, len(in))
	for _, p := range in {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

// ---------------------------------------------------------------
// main — wire the pipeline and run one question end-to-end.
// ---------------------------------------------------------------

const corpus = `# Retrieval-Augmented Generation

RAG combines retrieval with generation: the LLM is given documents fetched at query time, so it can answer questions about content it never saw at training time.

## Pipeline

Pipelines typically have four stages: parse, embed, store, and retrieve. Each step lives behind an interface so the pieces can be swapped without rewriting the orchestrator.

## Embeddings

Embeddings map text to dense vectors. Two vectors with high cosine similarity tend to mean similar things. Production embedders are hosted models; teaching codebases often use a deterministic hash instead.

---

## Vector stores

A vector store keeps (id, embedding, metadata) rows. The standard query is top-k nearest neighbors. In-memory linear scan suffices up to thousands of records; production uses pgvector, FAISS, or Qdrant.

## Prompt assembly

Once retrieval returns chunks, the prompt assembler concatenates them under a CONTEXT heading and instructs the LLM to answer ONLY from the context.
`

func main() {
	p := NewPipeline()
	n := p.Ingest("rag-mini", "rag-mini.md", corpus)
	fmt.Printf("ingested %d chunks into the store\n\n", n)

	for _, q := range []string{
		"What do embeddings do?",
		"What is in a vector store row?",
		"How does prompt assembly help?",
	} {
		ans, hits, prompt := p.Ask(q, 3)
		fmt.Printf("Q: %s\n", q)
		for i, h := range hits {
			fmt.Printf("  hit[%d] score=%+.3f id=%s\n", i+1, h.Score, h.Chunk.ID)
		}
		fmt.Printf("\n  prompt-len=%d chars\n", len(prompt))
		fmt.Printf("  A: %s\n\n", ans)
	}
}
