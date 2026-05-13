# s01 · block-model

## Problem

Every RAG framework eventually has to answer one quiet question:
"what *is* a piece of content, internally?" A PDF page can contain
running text, a table, an equation, a figure caption, and a code
listing. If your pipeline calls each of these `string`, you've already
lost — the chunker can't keep a table together, the embedder can't
choose a separate image-captioning path, and the retriever can't tell
the user "this answer came from a table on page 4". So before we parse
anything, we need a typed vocabulary.

## Solution

Chapter s01 establishes that vocabulary in `rag/types.go` and shows it
in action with a hand-built sample document. The Go types are:

- `BlockKind` — enum of `text | heading | code | image | table | equation`.
- `Block` — one typed unit with `Order`, `Page`, `Section`, `Meta`.
- `Document` — `ID`, `Source`, `[]Block`.
- (Stub for later) `Chunk`, `Embedding`, `VectorRecord`, `RetrievedChunk`.

The chapter's `main.go` constructs a 3-block sample document and prints
both a human-readable summary and the raw JSON, so you can see the data
shape with your own eyes before any algorithm shows up.

## How It Works

1. `buildSampleDoc` returns a `rag.Document` literal with three blocks:
   heading, text, code.
2. `summarize` walks the blocks and renders one line per block tagged
   by its `BlockKind`.
3. `main` prints the summary, then a `json.Encoder`-formatted dump of
   the same document so the JSON wire shape is visible.

There is intentionally no behavior in `rag/`: the package is types
only. Every later chapter adds its own concrete behavior in its own
directory.

## What Changed

This is chapter one — nothing changed yet. The point is to fix
vocabulary before anyone writes a parser, an embedder, or a retriever.

## Try It

```bash
go run ./agents/s01-block-model
```

Expected (truncated):

```
Document doc-001 (source=intro.md, 3 blocks)
  [heading] order=0 page=1 section="intro" text="RAG in one paragraph"
  [text] order=1 page=1 section="intro" text="Retrieval-Augmented Generation lets an LLM answer using d…"
  [code] order=2 page=1 section="intro" text="rag.Embed(query) -> top-k chunks -> prompt(LLM)"

Raw JSON:
{
  "ID": "doc-001",
  ...
}
```

Tests:

```bash
go test ./agents/s01-block-model
```

## Upstream Source Reading

The upstream framework wires modality into a long enum-style dispatch
inside `raganything/modalprocessors.py`, with `ImageModalProcessor`,
`TableModalProcessor`, `EquationModalProcessor`, and `GenericModalProcessor`
each handling its own block kind. The annotated excerpt under
[`upstream-readings/s01-blocks.go.md`](../../upstream-readings/s01-blocks.go.md)
walks through how a Python dataclass + `get_processor_for_type` lookup
plays the same role our `BlockKind` enum plays in Go.
