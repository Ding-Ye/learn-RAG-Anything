// Chapter s03 — chunker.
//
// Turns a stream of typed Blocks into Chunks suitable for embedding.
// The job of a chunker is to:
//
//   - hit a target size budget (so vectors are dense and embeddings
//     don't truncate),
//   - preserve some context across boundaries with a configurable
//     overlap (so retrieval doesn't slice mid-sentence),
//   - keep provenance (doc_id, section, pages) on every chunk.
//
// We use a simple character-count budget; real systems use token
// counts from the embedding-model tokenizer.
//
// Run: go run ./agents/s03-chunker
package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// ChunkConfig holds the policy knobs. Defaults are tiny on purpose so
// the demo produces several chunks from a short input.
type ChunkConfig struct {
	TargetChars  int // soft maximum size of a chunk
	OverlapChars int // characters re-used from the previous chunk
}

// Chunk produces ordered Chunks from a Document's Blocks. Headings and
// code blocks are kept as their own chunks (so they aren't split mid-
// fence); text blocks are concatenated into a sliding window.
func Chunk(doc rag.Document, cfg ChunkConfig) []rag.Chunk {
	if cfg.TargetChars <= 0 {
		cfg.TargetChars = 200
	}
	if cfg.OverlapChars < 0 || cfg.OverlapChars >= cfg.TargetChars {
		cfg.OverlapChars = cfg.TargetChars / 5
	}

	var (
		chunks    []rag.Chunk
		buf       strings.Builder
		bufPages  []int
		bufSec    string
		nextIndex int
	)

	flush := func() {
		text := strings.TrimSpace(buf.String())
		if text == "" {
			return
		}
		c := rag.Chunk{
			ID:      makeChunkID(doc.ID, nextIndex, text),
			Text:    text,
			DocID:   doc.ID,
			Section: bufSec,
			Pages:   dedupPages(bufPages),
			Meta:    map[string]string{"index": fmt.Sprintf("%d", nextIndex)},
		}
		chunks = append(chunks, c)
		nextIndex++

		// Seed the next buffer with an overlap tail so adjacent chunks
		// share some context.
		tail := text
		if cfg.OverlapChars > 0 && len(text) > cfg.OverlapChars {
			tail = text[len(text)-cfg.OverlapChars:]
		} else if cfg.OverlapChars == 0 {
			tail = ""
		}
		buf.Reset()
		bufPages = bufPages[:0]
		if tail != "" {
			buf.WriteString(tail)
		}
	}

	for _, b := range doc.Blocks {
		// Headings and code never get spliced into a running text
		// buffer; flush whatever is buffered, then emit them solo.
		if b.Kind == rag.BlockHeading || b.Kind == rag.BlockCode {
			flush()
			solo := rag.Chunk{
				ID:    makeChunkID(doc.ID, nextIndex, b.Text),
				Text:  b.Text,
				DocID: doc.ID,
				Section: func() string {
					if b.Kind == rag.BlockHeading {
						return b.Text
					}
					return b.Section
				}(),
				Pages: []int{b.Page},
				Meta: map[string]string{
					"index": fmt.Sprintf("%d", nextIndex),
					"kind":  string(b.Kind),
					"lang":  b.Lang,
				},
			}
			chunks = append(chunks, solo)
			nextIndex++
			bufSec = b.Section
			if b.Kind == rag.BlockHeading {
				bufSec = b.Text
			}
			continue
		}

		// Text block: append to the running buffer, flushing whenever
		// we'd cross the target size. If a single block is larger than
		// the target, split it into TargetChars-sized slices so no
		// chunk grows unbounded.
		bufSec = b.Section
		remaining := b.Text
		for len(remaining) > 0 {
			room := cfg.TargetChars - buf.Len()
			if room <= 0 {
				flush()
				continue
			}
			// Use whichever is smaller: what's left of the block, or
			// the remaining budget in the current chunk.
			take := room
			if take > len(remaining) {
				take = len(remaining)
			}
			if buf.Len() > 0 {
				buf.WriteByte(' ')
			}
			buf.WriteString(remaining[:take])
			remaining = remaining[take:]
			bufPages = append(bufPages, b.Page)
			if buf.Len() >= cfg.TargetChars {
				flush()
			}
		}
	}
	flush()
	return chunks
}

func makeChunkID(docID string, idx int, text string) string {
	h := sha1.Sum([]byte(docID + "|" + text))
	return fmt.Sprintf("%s-%03d-%s", docID, idx, hex.EncodeToString(h[:4]))
}

func dedupPages(in []int) []int {
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

func main() {
	doc := rag.Document{
		ID:     "demo",
		Source: "demo.md",
		Blocks: []rag.Block{
			{Kind: rag.BlockHeading, Text: "Intro", Page: 1, Order: 0, Section: "Intro"},
			{Kind: rag.BlockText, Text: strings.Repeat("RAG combines retrieval with generation. ", 6),
				Page: 1, Order: 1, Section: "Intro"},
			{Kind: rag.BlockText, Text: strings.Repeat("It is great for grounding LLM answers. ", 4),
				Page: 2, Order: 2, Section: "Intro"},
			{Kind: rag.BlockCode, Text: "rag.Embed(query)", Page: 2, Order: 3,
				Section: "Intro", Lang: "go"},
		},
	}
	chunks := Chunk(doc, ChunkConfig{TargetChars: 150, OverlapChars: 30})
	fmt.Printf("produced %d chunks (target=150, overlap=30)\n", len(chunks))
	for _, c := range chunks {
		fmt.Printf("  %s pages=%v section=%-8q len=%d\n     text=%q\n",
			c.ID, c.Pages, c.Section, len(c.Text), abbreviate(c.Text, 90))
	}
}

func abbreviate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
