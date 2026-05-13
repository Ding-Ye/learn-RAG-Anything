# s_full-integration · tracing one query end-to-end

## Problem

The chapters teach one stage at a time. That's the right shape for
learning, but it means the *seams* between stages are easy to lose. In
this doc we follow a single query — `"What do embeddings do?"` — all
the way from a markdown blob to a final LLM completion, naming every
piece of state at every stage and pointing at the file where it lives.

## Solution

A linear trace. No new code: we just narrate the run of `go run
./agents/s08-pipeline`, then explain what each chapter contributed and
where to find each step in the upstream Python source.

## How It Works

### Step 0 · the input

A markdown corpus (~600 chars, defined in `agents/s08-pipeline/main.go`
as `const corpus`) talks about RAG basics. Three short headings,
several paragraphs, one page break (`---`).

### Step 1 · parse (chapter s02)

`parseMarkdown(corpus)` scans the source line-by-line and emits
`[]rag.Block`. For our corpus you get something like:

```
[heading] "Retrieval-Augmented Generation"
[text]    "RAG combines retrieval with generation: ..."
[heading] "Pipeline"
[text]    "Pipelines typically have four stages: ..."
[heading] "Embeddings"
[text]    "Embeddings map text to dense vectors: ..."
[heading] "Vector stores"        (page=2 after the --- break)
[text]    "A vector store keeps (id, embedding, metadata) rows. ..."
[heading] "Prompt assembly"
[text]    "Once retrieval returns chunks, ..."
```

Each block carries its `Section` (last heading seen) and `Page`.

### Step 2 · chunk (chapter s03)

`chunkDoc(doc, {TargetChars:220, OverlapChars:30})` walks the blocks:

- Headings emit as their own (tiny) chunks. Result: `Retrieval-Augmented Generation`, `Pipeline`, `Embeddings`, ...
- Text blocks are spliced into a sliding window with 30-char overlap.

For our corpus that produces 12 chunks. Each chunk has:

- `ID` — `rag-mini-000-fb5d…` style, content-hashed.
- `DocID` — `rag-mini`.
- `Pages` — `[1]` or `[2]` (after the `---` break).
- `Section` — the relevant heading.
- `Text` — up to ~220 chars.

### Step 3 · embed (chapter s04)

For each chunk, `Pipeline.Ingest` calls `Embedder.Embed(chunk.Text)`
and gets back a 64-dim unit-length `rag.Embedding`. With the
`FakeEmbedder`, the vector is a deterministic SHA-256 scatter — two
chunks that share many tokens land near each other in cosine space.

### Step 4 · store (chapter s05)

Each `(chunkID, embedding, chunk)` triple is wrapped in
`rag.VectorRecord` and appended to the `MemoryStore`. After ingest:
`p.Store.Len() == 12`.

### Step 5 · ask: embed the query (chapter s04 again)

`Pipeline.Ask("What do embeddings do?", 3)` calls
`p.Embedder.Embed(question)` to produce a 64-dim query vector. *Crucially
the same embedder is used* — querying with a different embedder than
indexing makes scores meaningless.

### Step 6 · retrieve (chapters s05 + s06)

`p.Store.Search(qvec, 3)` linearly scans the 12 records, computes
`dot(qvec, rec.Embedding)` (cosine, since both sides are unit-length),
sorts descending, and returns the top 3.

For our query the top hit is the `Embeddings` heading-derived chunk —
exactly what you want when the user asks about embeddings.

### Step 7 · assemble the prompt (chapter s07)

`renderPrompt(question, hits)` produces:

```
You are a careful assistant. Answer the QUESTION using ONLY the CONTEXT.

CONTEXT:
[c1] Embeddings
[c2] (...next-best chunk text...)
[c3] (...third-best chunk text...)

QUESTION:
What do embeddings do?

ANSWER:
```

The `[c1] [c2] [c3]` markers let the LLM cite by index.

### Step 8 · LLM completion (chapter s08)

`MockLLM.Complete(prompt)` parses the `CONTEXT:` block out of the
prompt and returns a synthetic answer that quotes `[c1]`. In a real
deployment this is replaced with an OpenAI / Anthropic / local-model
client; the interface (`Complete(prompt) string`) stays the same.

## What Changed

This doc adds no new code. It freezes the conceptual map of the
curriculum:

| Stage     | Chapter | Output type                |
| --------- | ------- | -------------------------- |
| Parse     | s02     | `[]rag.Block`              |
| Chunk     | s03     | `[]rag.Chunk`              |
| Embed     | s04     | `rag.Embedding` per chunk  |
| Store     | s05     | `rag.VectorRecord` rows    |
| Retrieve  | s06     | `[]rag.RetrievedChunk`     |
| Assemble  | s07     | a single prompt string     |
| Generate  | s08     | the answer string          |

## Try It

```bash
go run ./agents/s08-pipeline
```

Watch the printed chunk IDs change as you tweak the corpus or the
chunk config in `agents/s08-pipeline/main.go`.

## Upstream Source Reading

A query that exercises the same path through the upstream framework
hits, in order:

| Stage     | Upstream file                                   |
| --------- | ----------------------------------------------- |
| Parse     | `raganything/parser.py` (MinerU/Docling)        |
| Chunk     | `raganything/processor.py` (`_process_text_chunks`) |
| Embed     | injected `embedding_func`, called by LightRAG    |
| Store     | `lightrag.storage.*VectorDBStorage`              |
| Retrieve  | `raganything/query.py` + LightRAG `aquery`       |
| Assemble  | `raganything/prompt.py` + `prompt_manager.py`    |
| Generate  | injected `llm_model_func`                        |

The companion appendix `appendix-b-upstream-map.md` lists every
upstream file we cared about and what it owns.
