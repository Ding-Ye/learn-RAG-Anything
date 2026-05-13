# Upstream reading · pipeline (s08)

> `.go.md` extension so `go build` ignores this file.

**Source**: `raganything/raganything.py` from
`HKUDS/RAG-Anything @ 146828f7…` (MIT).

## The orchestrator

Upstream RAGAnything is a multi-inheritance class that fuses three
mixins:

```python
# raganything/raganything.py  (sketch, comments added)

class RAGAnything(QueryMixin, ProcessorMixin, BatchMixin):
    """Top-level orchestrator. Each mixin owns one concern."""

    def __init__(
        self,
        working_dir: str,
        # ... config ...
        llm_model_func: Callable | None = None,
        vision_model_func: Callable | None = None,
        embedding_func: EmbeddingFunc | None = None,
    ):
        # Build the underlying LightRAG with injected callables.
        self.lightrag = LightRAG(
            working_dir=working_dir,
            embedding_func=embedding_func,
            llm_model_func=llm_model_func,
            ...
        )
        # Modal processor cache, parser config, callbacks.
        ...

    async def process_document_complete(self, file_path: str, **kwargs):
        # 1) Parse → list of typed blocks (Parser).
        # 2) Process blocks (ProcessorMixin):
        #      - text → chunk → embed → upsert into vector store
        #      - image / table / equation → modal processor → caption + embed
        # 3) Build/refresh the entity-relation graph.
        ...

    async def aquery(self, query: str, mode: str = "hybrid", **kwargs):
        # Delegated to LightRAG via QueryParam(mode=...)
        ...

    async def aquery_with_multimodal(self, query, multimodal_content, **kwargs):
        # Routes images / tables in the query through their modal
        # processors before retrieving.
        ...
```

## Key observations

1. **Composition by mixin.** Each concern (process / query / batch)
   gets its own class; Python multiple-inheritance assembles the
   orchestrator. In Go we replicate this by embedding interfaces in
   a struct — same idea, different syntax.
2. **One construction, many methods.** A single object is configured
   once and answers both ingest and query calls; nothing is reloaded
   between requests.
3. **Everything callable is injected.** `llm_model_func`,
   `vision_model_func`, `embedding_func` come in via `__init__`. Our
   Go pipeline takes `Embedder`, `VectorStore`, `LLM` as interface
   fields — directly equivalent.
4. **Multimodal is a parallel path, not a special case.** Upstream
   keeps `aquery` text-only and `aquery_with_multimodal` for the rest.
   The shape matches: most users will only need the text path; the
   multi-modal extension reuses the same retrieve+prompt machinery.

## Map to our Go types

| Upstream                                              | This repo                                |
| ----------------------------------------------------- | ---------------------------------------- |
| `RAGAnything(QueryMixin, ProcessorMixin, BatchMixin)` | `type Pipeline struct {…}`               |
| `process_document_complete(file_path)`                | `Pipeline.Ingest(docID, source, md)`     |
| `aquery(query, mode="hybrid")`                        | `Pipeline.Ask(question, k)`              |
| `llm_model_func`                                      | `LLM` interface + `MockLLM`              |
| `embedding_func`                                      | `Embedder` interface                     |
| `vision_model_func`                                   | not implemented — see Appendix A         |
| `BatchMixin`                                          | out of scope                             |

## Where to look in upstream

| Concept                  | File                          | Hint                                |
| ------------------------ | ----------------------------- | ----------------------------------- |
| Class declaration        | `raganything/raganything.py`  | `class RAGAnything(...)`            |
| `process_document_complete` | `raganything/processor.py` | search by name (it's on the mixin)  |
| `aquery` / multimodal    | `raganything/query.py`        | `class QueryMixin`                  |
| Batch flow               | `raganything/batch.py`        | `class BatchMixin`                  |
| Configuration            | `raganything/config.py`       | `RAGAnythingConfig`                 |
