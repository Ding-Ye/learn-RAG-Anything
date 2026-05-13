// Chapter s01 — block model.
//
// Goal: introduce the typed vocabulary every later chapter shares.
// We build a tiny in-memory Document by hand, then print a flat
// summary that shows how a Block carries its modality and metadata.
//
// Run: go run ./agents/s01-block-model
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// buildSampleDoc constructs a tiny Document by hand. In later chapters
// this comes from a parser; here we want the reader to *see* the data
// shape first, before any parser code shows up.
func buildSampleDoc() rag.Document {
	return rag.Document{
		ID:     "doc-001",
		Source: "intro.md",
		Blocks: []rag.Block{
			{
				Kind:    rag.BlockHeading,
				Text:    "RAG in one paragraph",
				Page:    1,
				Order:   0,
				Section: "intro",
			},
			{
				Kind:    rag.BlockText,
				Text:    "Retrieval-Augmented Generation lets an LLM answer using documents it never saw at training time.",
				Page:    1,
				Order:   1,
				Section: "intro",
			},
			{
				Kind:    rag.BlockCode,
				Text:    "rag.Embed(query) -> top-k chunks -> prompt(LLM)",
				Page:    1,
				Order:   2,
				Section: "intro",
				Lang:    "pseudo",
			},
		},
	}
}

// summarize is the "feature" of this chapter: walking the typed Blocks
// and producing a stable, human-readable trace. It exists so this
// program does something observable beyond a struct literal.
func summarize(doc rag.Document) []string {
	out := make([]string, 0, len(doc.Blocks))
	for _, b := range doc.Blocks {
		line := fmt.Sprintf("[%s] order=%d page=%d section=%q text=%q",
			b.Kind, b.Order, b.Page, b.Section, truncate(b.Text, 60))
		out = append(out, line)
	}
	return out
}

// truncate keeps the demo output friendly to a terminal width without
// destroying the original Block contents.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func main() {
	doc := buildSampleDoc()

	fmt.Printf("Document %s (source=%s, %d blocks)\n",
		doc.ID, doc.Source, len(doc.Blocks))
	for _, line := range summarize(doc) {
		fmt.Println("  " + line)
	}

	// Also dump the raw JSON so the data shape is unambiguous.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	fmt.Println("\nRaw JSON:")
	if err := enc.Encode(doc); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}
}
