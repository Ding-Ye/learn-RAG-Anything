# s03 · chunker

## Problem

Embedders have a maximum input size (often 512 / 8K tokens). A whole
parsed document is far too large to embed as one vector — even if it
weren't, you'd lose the ability to retrieve *which paragraph* answered
a question. So we need to *chunk*: split a Document's Blocks into
overlapping, embeddable pieces while keeping enough metadata that a
retrieved chunk still tells us the doc, section, and page it came from.

## Solution

`Chunk(doc, cfg)` walks the Block stream and accumulates text into a
running buffer until it reaches `cfg.TargetChars`, at which point it
flushes a `rag.Chunk`. Two policies make the output friendlier for
retrieval:

- **Headings and code blocks bypass the buffer.** A heading becomes
  its own (tiny) chunk so it can be matched directly; a code block
  stays whole so we never fragment a function across two embeddings.
- **Overlap.** Each flush seeds the next chunk with the last
  `cfg.OverlapChars` characters, so a query that lives on a chunk
  boundary still has a chance to retrieve a chunk containing both
  sides of the boundary.

Provenance — `DocID`, `Section`, `Pages` — rides along on every chunk.

## How It Works

1. `flush()` is the central operation: take the running buffer's text,
   build a `rag.Chunk` (id derived from `sha1(docID|text)`), record its
   pages, then re-seed the buffer with an overlap tail.
2. The main loop has two paths:
   - heading / code → flush, then emit the block as a solo chunk.
   - text → append into the buffer, splitting *inside* the block if a
     single block is larger than the target so no chunk grows unbounded.
3. `dedupPages` collapses page numbers (a chunk spanning pages 2-3
   shouldn't list 2 twice).

The chunker has no knowledge of embeddings, retrieval, or LLMs — its
only job is `(Document, ChunkConfig) → []Chunk`.

## What Changed

- Added `rag.Chunk` users: this is the first chapter that produces them.
- Introduced `ChunkConfig` so the policy is explicit and tunable
  without rewriting the function.

## Try It

```bash
go run ./agents/s03-chunker
```

Expected (truncated):

```
produced 7 chunks (target=150, overlap=30)
  demo-000-17e341b9 pages=[1] section="Intro"  len=5    text="Intro"
  demo-001-96c598c0 pages=[1] section="Intro"  len=150  text="RAG combines retrieval ..."
  demo-002-d076f332 pages=[1 2] section="Intro" len=151 text="...generation. RAG combines..."
  ...
```

Notice how the heading "Intro" gets its own tiny chunk and the code
chunk `rag.Embed(query)` is kept whole.

Tests:

```bash
go test ./agents/s03-chunker
```

## Upstream Source Reading

The upstream chunker lives inside `raganything/processor.py` and is
called as part of the bigger `process_document_complete` flow. It uses
LightRAG's tokenizer for the size budget rather than naive character
counts, and stores chunk records via `compute_mdhash_id`. See
[`upstream-readings/s03-chunker.go.md`](../../upstream-readings/s03-chunker.go.md)
for the relevant excerpt and the mapping to our Go implementation.
