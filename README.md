# learn-RAG-Anything

A from-zero, hands-on Go companion to
[HKUDS/RAG-Anything](https://github.com/HKUDS/RAG-Anything) — eight short
chapters that build a tiny but real Retrieval-Augmented Generation (RAG)
pipeline one mechanism at a time.

Each chapter is a self-contained Go program (~100 LOC) plus a bilingual
explainer and a guided reading of the corresponding upstream Python
source. You can `go run ./agents/sNN-*` for any chapter individually.

**中文版**: see [README.zh.md](./README.zh.md).

## Curriculum

| #   | Chapter             | What you build                                    | Run                                |
| --- | ------------------- | ------------------------------------------------- | ---------------------------------- |
| s01 | block-model         | Typed Block / Chunk vocabulary                    | `go run ./agents/s01-block-model`  |
| s02 | parser              | Tiny Markdown → []Block parser                    | `go run ./agents/s02-parser`       |
| s03 | chunker             | Blocks → overlapping Chunks with metadata         | `go run ./agents/s03-chunker`      |
| s04 | embedder            | Deterministic hash-based Embedder (no model)      | `go run ./agents/s04-embedder`     |
| s05 | vector-store        | In-memory store + cosine-similarity top-k         | `go run ./agents/s05-vector-store` |
| s06 | retriever           | Query → embed → top-k chunks                      | `go run ./agents/s06-retriever`    |
| s07 | prompt-assembler    | Chunks + question → LLM prompt                    | `go run ./agents/s07-prompt-assembler` |
| s08 | pipeline            | End-to-end parse → chunk → embed → store → answer | `go run ./agents/s08-pipeline`     |

Plus bilingual documentation under [`docs/en`](./docs/en) and
[`docs/zh`](./docs/zh):

- End-to-end trace: [`s_full-integration`](./docs/en/s_full-integration.md)
- Appendix A: [Multi-modal RAG](./docs/en/appendix-a-multimodal-rag.md) ([中文](./docs/zh/appendix-a-multimodal-rag.md))
- Appendix B: [Upstream file map](./docs/en/appendix-b-upstream-map.md) ([中文](./docs/zh/appendix-b-upstream-map.md))

## Repository layout

```
rag/                     shared, behavior-free types (Block, Chunk, ...)
agents/sNN-<name>/       one chapter = one main.go + tests + README
docs/en/, docs/zh/       bilingual chapter explainers, six-section format
upstream-readings/       annotated upstream-source excerpts (.go.md)
web/index.html           tiny static curriculum viewer
.github/workflows/       CI: vet, build, test, bilingual heading parity
```

## Quick start

```bash
git clone https://github.com/Ding-Ye/learn-RAG-Anything.git
cd learn-RAG-Anything
go test ./...
go run ./agents/s08-pipeline
```

## Six-section doc format

Every chapter doc follows the same shape, so you always know where to
look:

1. **Problem** — what real-world question motivates this stage.
2. **Solution** — how this chapter answers it in ~100 lines of Go.
3. **How It Works** — the moving parts and data flow.
4. **What Changed** — what's new since the previous chapter.
5. **Try It** — concrete commands and expected output.
6. **Upstream Source Reading** — pointer into the matching Python file.

## License & acknowledgement

This repository is licensed under **Apache License 2.0**
(see [LICENSE](./LICENSE)).

It is a re-implementation for educational purposes, inspired by
[HKUDS/RAG-Anything](https://github.com/HKUDS/RAG-Anything) at commit
`146828f73de652c9d72399bfc60499966f3f8bd0`, which is licensed under MIT.
We gratefully acknowledge the upstream authors. No upstream code is
copied verbatim into this repo; the `upstream-readings/` directory
contains short, commented excerpts under fair-use academic citation.

## A note on scope

This is a **teaching** repo, not a production RAG framework. The
embedder is a hash; the vector store is an in-memory slice; the LLM is
mocked. Every shortcut is documented in the chapter docs, so when you
later read the real upstream you know which abstractions are realistic
and which are stand-ins.
