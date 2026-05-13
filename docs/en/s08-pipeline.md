# s08 · pipeline

## Problem

Each of s02..s07 demonstrates *one* mechanism. The natural next step
is to put them together: feed a markdown document into the front
of the pipeline, ask a question at the back, and watch retrieval
guide the answer. This is the chapter where the abstract gets concrete.

It also forces a design choice: should the orchestrator know each
stage's internals, or treat each as a plug behind an interface? The
answer in real systems (and in this chapter) is the latter — the
pipeline is small precisely because every stage hides its logic.

## Solution

`Pipeline` has three fields:

```go
type Pipeline struct {
    Embedder Embedder
    Store    VectorStore
    LLM      LLM
}
```

Two methods:

- `Ingest(docID, source, markdown) int` — parse, chunk, embed, store.
- `Ask(question, k) (answer, hits, prompt)` — embed, retrieve,
  assemble, complete.

To keep the chapter self-contained, the prior stages are re-declared
as small private helpers inside the same `package main` — each is a
faithful shrink of its chapter, with comments pointing back. Replacing
any of them with the full chapter implementation or a production
backend is straightforward because the interfaces (`Embedder`,
`VectorStore`, `LLM`) are tiny.

## How It Works

`main` exercises the pipeline against a short, realistic markdown
corpus that mirrors the kind of doc RAG systems actually consume:

1. `NewPipeline()` wires a 64-dim hash embedder, an in-memory store,
   and a `MockLLM`. The mock LLM doesn't call out — it parses
   `CONTEXT:` from the prompt and echoes the top chunk so you can see
   the answer comes from retrieval, not from somewhere else.
2. `Ingest("rag-mini", "rag-mini.md", corpus)` runs the four-stage
   indexing pipeline once and reports how many chunks landed.
3. `Ask(q, 3)` for several questions: for each, it prints the top-3
   chunk ids and the assembled prompt's length, then the synthetic
   answer that quotes `[c1]`.

The flow is exactly the one upstream documents:

```
markdown  → parser → []Block → chunker → []Chunk
                                      ↓
                                  embedder
                                      ↓
                                  VectorStore
question → embedder → Search → []RetrievedChunk → prompt → LLM → answer
```

## What Changed

- First chapter that orchestrates the previous mechanisms.
- Adds the `LLM` interface and `MockLLM` so the pipeline is runnable
  without any external API key.
- Demonstrates the *interchangeability* of stages by accepting them
  as interfaces.

## Try It

```bash
go run ./agents/s08-pipeline
```

Expected (truncated):

```
ingested 12 chunks into the store

Q: What do embeddings do?
  hit[1] score=+0.452 id=rag-mini-004-fb5d5812
  hit[2] score=+0.126 id=rag-mini-003-e12ca782
  hit[3] score=+0.121 id=rag-mini-010-29d646e7

  prompt-len=373 chars
  A: (mock answer) Based on the top-ranked chunk, [c1] Embeddings ...
```

Tests:

```bash
go test ./agents/s08-pipeline
```

For a written walk-through of one query end-to-end (with citations
back to the upstream files), see
[`docs/en/s_full-integration.md`](./s_full-integration.md).

## Upstream Source Reading

Upstream orchestration lives in `raganything/raganything.py`, with
`process_document_complete` (ingest) and `aquery` / `aquery_with_multimodal`
(ask) sitting at the top of a stack of mixins (`ProcessorMixin`,
`QueryMixin`, `BatchMixin`). See
[`upstream-readings/s08-pipeline.go.md`](../../upstream-readings/s08-pipeline.go.md)
for the relevant excerpt — note how the mixin-per-concern pattern in
Python plays the role of our "compose interfaces in a struct" pattern.
