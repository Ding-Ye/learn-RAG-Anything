# Upstream reading · retriever (s06)

> `.go.md` extension so `go build` ignores this file.

**Source**: `raganything/query.py` from
`HKUDS/RAG-Anything @ 146828f7…` (MIT).

## What the upstream retriever actually does

```python
# raganything/query.py  (sketch, comments added)

from lightrag import QueryParam

class QueryMixin:
    async def aquery(self, query: str, mode: str = "hybrid", **kwargs):
        # 1) Build a LightRAG QueryParam from mode + knobs.
        param = QueryParam(mode=mode, top_k=kwargs.get("top_k", 60), ...)

        # 2) LightRAG runs the actual retrieval.
        #    Depending on `mode`, this can:
        #      - "naive"  : vector similarity over chunks only
        #      - "local"  : entity-graph local subgraph
        #      - "global" : entity-graph global summary
        #      - "hybrid" : local + global blended
        #      - "mix"    : graph + vector chunks blended
        return await self.lightrag.aquery(query, param=param)
```

Multi-modal queries follow a parallel path (`aquery_with_multimodal`):
images / tables / equations attached to the query are routed through
their own modal processors which may caption / linearize / render
before retrieval kicks in.

## Key observations

1. **The retriever picks a *mode*, not just a top-k.** Real systems
   blend graph and vector retrieval. We model only the `naive` /
   chunk-vector mode.
2. **`top_k` defaults are big (often 60).** Real retrievers over-fetch
   heavily and rely on reranking + truncation. Our `overFetch := 3*K`
   captures the same instinct at a small scale.
3. **Query rewriting happens before this layer.** Upstream may call an
   LLM to extract entities or rewrite the query into multiple
   sub-queries; we keep that out of scope.
4. **Async by default.** Network calls to embedders / LLMs / vector
   stores make sync code dangerous. Our Go version is sync; the
   pattern translates to `context.Context`-aware functions if you
   later go async.

## Map to our Go types

| Upstream                                | This repo                                |
| --------------------------------------- | ---------------------------------------- |
| `QueryParam.mode = "naive"`             | the only mode we implement                |
| `QueryParam.top_k = 60`                 | `RetrieveOptions.K` (default 5)          |
| `lightrag.aquery(...)`                  | `Retriever.Retrieve(question, opts)`     |
| Modal query branch (image / table / …)  | Appendix A material                       |
| Graph retrieval branch                  | not implemented                           |
| Async                                   | sync (teaching-scale)                     |

## Where to look in upstream

| Concept                | File                       | Hint                              |
| ---------------------- | -------------------------- | --------------------------------- |
| `QueryMixin.aquery`    | `raganything/query.py`     | search `def aquery`               |
| `QueryParam`           | `lightrag` (external)      | `from lightrag import QueryParam` |
| Multi-modal query path | `raganything/query.py`     | search `aquery_with_multimodal`   |
| Cache key for queries  | `raganything/query.py`     | `_generate_multimodal_cache_key`  |
