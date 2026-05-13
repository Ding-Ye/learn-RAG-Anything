# s04 · embedder

## Problem

Retrieval needs a comparable form of "meaning". The standard answer is
an *embedding*: a dense vector where semantically similar inputs land
near each other. In real systems this is an API call to a hosted model,
which costs money, requires network, and ties tests to credentials.
For a teaching repo we need something that:

- never makes a network call (CI must pass offline),
- gives the same vector for the same input on every machine,
- still produces *useful* similarity scores so the next chapter's
  retrieval demo isn't all noise.

## Solution

Define `Embedder` as a tiny Go interface:

```go
type Embedder interface {
    Dim() int
    Embed(text string) rag.Embedding
}
```

Then implement `FakeEmbedder`: tokenize on whitespace + punctuation,
SHA-256 each token, scatter the bytes of each hash into a fixed-size
vector with +/- contributions, and L2-normalize the result. Two inputs
that share many tokens land near each other in cosine space — not
because the model "understands" them, but because they map to
overlapping positions.

The interface is what the rest of the pipeline depends on. Swapping in
a real model later is a one-file change.

## How It Works

For each token `t`:

1. `sum = SHA256(t)` — 32 bytes of deterministic noise.
2. Walk `sum` in 4-byte windows. Each window picks a position `p` in
   `[0, dim)` and a +/- sign from the top bit.
3. Add `±1/len(tokens)` to `vec[p]`. Dividing by token count keeps the
   magnitudes comparable across inputs of different lengths.
4. After all tokens: `l2Normalize(vec)` so the cosine of two embeddings
   is just their dot product.

`CosineSimilarity` is exposed here as a small helper; s05 uses it for
top-k search.

## What Changed

- First non-trivial behavior on `rag.Embedding`.
- Introduced an interface boundary — every later chapter takes an
  `Embedder`, not a concrete type, so the fake can be swapped for a
  real model without touching the rest of the code.

## Try It

```bash
go run ./agents/s04-embedder
```

Expected:

```
Embedder dim=64
  sim=+0.430  "retrieval augmented generation"  vs  "RAG combines retrieval with generation"
  sim=-0.120  "retrieval augmented generation"  vs  "the cat sat on the mat"
  sim=+1.000  "hello world"  vs  "hello world"
```

The matching pair scores >0; the unrelated pair scores ~0 or slightly
negative; identical inputs score 1.0. That's exactly the property
retrieval needs.

Tests:

```bash
go test ./agents/s04-embedder
```

## Upstream Source Reading

The upstream framework treats the embedder as an *injected callable*:
`raganything/raganything.py` accepts an `embedding_func` and forwards
it to LightRAG, which then calls it from inside its chunking + indexing
loop. The model itself is whatever the user wires up (OpenAI, etc.).
See [`upstream-readings/s04-embedder.go.md`](../../upstream-readings/s04-embedder.go.md)
for the relevant excerpt and a side-by-side with our `Embedder` interface.
