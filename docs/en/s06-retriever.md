# s06 · retriever

## Problem

A user asks a question in natural language; the vector store speaks
vectors and chunks. Bridging the two is the retriever's job. In real
systems, the retriever is also where pre/post-processing lives: query
rewriting, hybrid (BM25 + vector) search, metadata filtering,
re-ranking, MMR for diversity. We isolate the *minimal* retriever from
those layers so the contract is visible.

## Solution

A small `Retriever` struct that holds an `Embedder` and a
`VectorStore`. `Retrieve(question, opts)`:

1. embeds the question with the *same* embedder used at indexing time
   (otherwise scores are meaningless),
2. asks the store for top-K similar records (over-fetched slightly to
   leave room for filtering),
3. applies optional `SectionFilter`, `MetaFilter`, and `MinScore`
   filters,
4. returns up to `K` `RetrievedChunk` records sorted by score.

## How It Works

```go
type RetrieveOptions struct {
    K            int
    SectionFilter string
    MinScore     float32
    MetaFilter   map[string]string
}
```

The two ingredients we depend on (`Embedder`, `VectorStore`) are
interfaces, not concrete types: in production you'd pass an OpenAI
embedder and a pgvector store; in tests you pass the s04 fake and the
s05 memory store. The retriever itself doesn't need to change.

Over-fetching matters: if `K=5` and we apply a section filter, we may
discard most of the top-5. We pull `3*K` (minimum 10) so filters have
budget. This is exactly what production retrievers do, just with much
larger numbers.

## What Changed

- First chapter that *composes* earlier chapters: imports nothing new
  from `rag/`, only re-declares the interfaces it needs.
- Introduces filter semantics on the retrieval boundary, leaving the
  store ignorant of metadata details.

## Try It

```bash
go run ./agents/s06-retriever
```

Expected (truncated):

```
Q: "How does retrieval work?"
  #1 score=+0.291 section="Pipeline" text="Retrieval finds top-k similar chunks"
  #2 score=+0.188 section="Pipeline" text="RAG combines retrieval with generation"
  ...

Q: "blocks" (section=Parsing)
  #1 score=+0.502 text="Markdown headings become heading blocks"
  #2 score=+0.000 text="Page numbers travel with each block"
```

Tests:

```bash
go test ./agents/s06-retriever
```

## Upstream Source Reading

Upstream retrieval lives in `raganything/query.py`, but most of the
heavy lifting happens inside LightRAG. The interesting piece in
`query.py` is *modal* retrieval: in addition to text chunks, the
framework can pull from an entity graph, with several modes
(`naive`, `local`, `global`, `hybrid`, `mix`) defined via
`lightrag.QueryParam`. See
[`upstream-readings/s06-retriever.go.md`](../../upstream-readings/s06-retriever.go.md)
for the relevant excerpt and how it maps to our single-mode Go retriever.
