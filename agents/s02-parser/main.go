// Chapter s02 — markdown parser.
//
// Turns a markdown string into a stream of typed Blocks. We support
// three shapes that span the interesting cases for retrieval:
//
//   - ATX headings ("# title", "## subtitle")
//   - fenced code blocks ("```lang ... ```")
//   - everything else groups into text blocks at paragraph boundaries
//
// The real upstream parser dispatches to MinerU / Docling / PaddleOCR
// to extract from PDFs and Office files. Here we strip the I/O and
// focus on the *transformation* step: bytes -> []rag.Block.
//
// Run: go run ./agents/s02-parser
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// ParseMarkdown converts a markdown source into a slice of Blocks.
// The function is deliberately small: it is a teaching artifact, not
// a CommonMark-compliant parser.
func ParseMarkdown(src string) []rag.Block {
	lines := strings.Split(src, "\n")
	blocks := make([]rag.Block, 0, 8)

	var (
		section   string
		paragraph []string
		page      = 1
		order     int
	)

	// flushParagraph emits the buffered text lines as one text block.
	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		blocks = append(blocks, rag.Block{
			Kind:    rag.BlockText,
			Text:    strings.TrimSpace(strings.Join(paragraph, " ")),
			Page:    page,
			Order:   order,
			Section: section,
		})
		order++
		paragraph = paragraph[:0]
	}

	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		trim := strings.TrimSpace(ln)

		// Page break marker: a horizontal rule "---" on its own line.
		// In real parsers, page numbers come from the upstream PDF
		// extractor; we expose the concept with a tiny in-band marker.
		if trim == "---" {
			flushParagraph()
			page++
			continue
		}

		// ATX heading.
		if strings.HasPrefix(trim, "#") {
			flushParagraph()
			level := 0
			for level < len(trim) && trim[level] == '#' {
				level++
			}
			title := strings.TrimSpace(trim[level:])
			section = title
			blocks = append(blocks, rag.Block{
				Kind:    rag.BlockHeading,
				Text:    title,
				Page:    page,
				Order:   order,
				Section: section,
				Meta:    map[string]string{"level": fmt.Sprintf("%d", level)},
			})
			order++
			continue
		}

		// Fenced code block.
		if strings.HasPrefix(trim, "```") {
			flushParagraph()
			lang := strings.TrimSpace(strings.TrimPrefix(trim, "```"))
			var body []string
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				body = append(body, lines[i])
				i++
			}
			blocks = append(blocks, rag.Block{
				Kind:    rag.BlockCode,
				Text:    strings.Join(body, "\n"),
				Page:    page,
				Order:   order,
				Section: section,
				Lang:    lang,
			})
			order++
			continue
		}

		// Blank line ends the current paragraph.
		if trim == "" {
			flushParagraph()
			continue
		}

		paragraph = append(paragraph, trim)
	}
	flushParagraph()
	return blocks
}

const sampleDoc = `# RAG-Anything (mini)

Retrieval-Augmented Generation lets an LLM answer using documents it never saw at training time.

## Pipeline

The pipeline has four stages: parse, embed, retrieve, generate.

` + "```go\n" + `chunks := store.Search(rag.Embed(query), 5)
` + "```\n" + `
---

## Multi-modal

Tables, images and equations need their own handlers; that's the topic of chapter s03 onward.
`

func main() {
	blocks := ParseMarkdown(sampleDoc)
	fmt.Printf("parsed %d blocks\n", len(blocks))
	for _, b := range blocks {
		fmt.Printf("  [%s] page=%d section=%-12q lang=%-3s text=%q\n",
			b.Kind, b.Page, b.Section, b.Lang, truncate(b.Text, 60))
	}
	if _, err := os.Stdout.WriteString("\n"); err != nil {
		os.Exit(1)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
