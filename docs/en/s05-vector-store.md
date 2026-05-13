# s05 · vector-store

## Problem

Once we have embeddings, we need somewhere to *put* them — and a way to
ask "give me the k chunks whose vectors are most similar to this
query". In production this is FAISS / pgvector / Qdrant / Pinecone /
Milvus; the storage technology differs but the contract is the same:
`Add(record)` and `Search(query, k) -> top-k`. Mistaking that contract
for the storage technology is a common confusion — this chapter
isolates it.

## Solution

Define a `VectorStore` interface and ship one trivial implementation,
`MemoryStore`, that linearly scans an in-memory slice. Both fit on
half a screen, which is exactly the point: the contract is small and
the storage is interchangeable.

```go
type VectorStore interface {
    Add(rec rag.VectorRecord)
    Search(query rag.Embedding, k int) []rag.RetrievedChunk
    Len() int
}
```

Search uses cosine similarity, computed as a dot product because s04's
embeddings are unit-length.

## How It Works

1. `Add(rec)` appends a `rag.VectorRecord` (chunk id, embedding, the
   chunk itself) to the slice.
2. `Search(query, k)` scores every record by `dot(query, rec.Embedding)`,
   sorts descending, and returns the first `k`.
3. The chapter includes a self-contained mini-embedder (FNV hash, two
   Newton iterations for sqrt) so the demo runs with zero imports
   outside of `rag`. Real code uses the s04 Embedder.

For a teaching repo, linear scan is the right answer:

- complexity is `O(n*dim)` per query, transparent and tiny,
- no buckets, no clusters, no ANN approximations,
- when you later move to FAISS you will *re-encode* this contract, not
  rewrite the pipeline.

## What Changed

- First chapter that produces `[]rag.RetrievedChunk` — the type that
  s06 and s07 consume.
- First interface boundary on storage; the rest of the curriculum
  takes a `VectorStore`, not a concrete `MemoryStore`.

## Try It

```bash
go run ./agents/s05-vector-store
```

Expected:

```
indexed 5 chunks (dim=32)

query: "retrieval and generation"
  #1 score=+0.612 id=a text="RAG combines retrieval with text generation"
  #2 score=+0.236 id=e text="Retrieval picks top-k chunks by similarity"
  ...
```

Tests:

```bash
go test ./agents/s05-vector-store
```

## Upstream Source Reading

In upstream the vector store is *not* implemented in `raganything/` —
it's owned by the LightRAG dependency. The interface RAGAnything sees
is the LightRAG `VectorDBStorage` family (one for chunks, one for
entities, one for relations). See
[`upstream-readings/s05-vector-store.go.md`](../../upstream-readings/s05-vector-store.go.md)
for how RAGAnything wires those storages in `raganything.py`'s
`__init__`, and how the contract maps to our small Go interface.
