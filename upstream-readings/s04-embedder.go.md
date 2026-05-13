# Upstream reading · embedder (s04)

> `.go.md` extension so `go build` ignores this file.

**Source**: `raganything/raganything.py` from
`HKUDS/RAG-Anything @ 146828f7…` (MIT).

The upstream framework doesn't *implement* an embedder — it accepts one
as a callable. The top-level orchestrator (`RAGAnything`) lets the user
plug in `llm_model_func`, `vision_model_func`, and `embedding_func`.

```python
# raganything/raganything.py  (sketch, comments added)

class RAGAnything(QueryMixin, ProcessorMixin, BatchMixin):
    def __init__(
        self,
        working_dir: str,
        # ... config ...
        llm_model_func: Callable | None = None,
        vision_model_func: Callable | None = None,
        embedding_func: EmbeddingFunc | None = None,
    ):
        # The embedder is wired into LightRAG, which then calls it
        # during chunk indexing.
        self.lightrag = LightRAG(
            working_dir=working_dir,
            embedding_func=embedding_func,
            llm_model_func=llm_model_func,
            ...
        )
```

`EmbeddingFunc` from `lightrag.utils` is roughly:

```python
@dataclass
class EmbeddingFunc:
    embedding_dim: int
    max_token_size: int
    func: Callable[[List[str]], Awaitable[List[List[float]]]]
```

So in upstream the interface is essentially:

> "Give me a function that takes a batch of strings and returns a batch
> of vectors, plus its dim and max input."

## Key observations

1. **The embedder is dependency-injected.** Upstream stays model-
   agnostic; the user decides whether to use OpenAI, Voyage, a local
   sentence-transformer, etc.
2. **Batched is the default shape.** `func` takes `List[str]` and
   returns `List[List[float]]` to amortize HTTP overhead.
3. **Dim and max-tokens travel with the function.** Downstream code
   needs both — to allocate matrices and to ensure chunks fit.
4. **Async by default.** The function is awaitable so a slow remote
   call doesn't block the event loop.

## Map to our Go types

| Upstream                                  | This repo                          |
| ----------------------------------------- | ---------------------------------- |
| `EmbeddingFunc.embedding_dim`             | `Embedder.Dim()`                   |
| `EmbeddingFunc.func`                      | `Embedder.Embed(text)` (single)    |
| `embedding_func=` constructor parameter   | accept an `Embedder` interface     |
| `vision_model_func=`                      | not implemented — see Appendix A   |
| Async / batched                           | sync / single (teaching simplicity)|

## What the fake hides

Real embedders learn the geometry of language. Our hash-based fake
gives you a *deterministic* but *content-aware* score: two strings that
share tokens land at overlapping coordinates, and the resulting
cosine reflects token overlap, not semantic similarity. That is enough
for the retrieval demo in s05/s06 to show meaningful ranking, but
nothing about synonymy, paraphrase, or multilingual matching survives.

## Where to look in upstream

| Concept                | File                                  | Hint                  |
| ---------------------- | ------------------------------------- | --------------------- |
| `embedding_func` param | `raganything/raganything.py`          | `__init__` signature  |
| `EmbeddingFunc` dataclass | `lightrag.utils` (external dep)    | not in this repo      |
| Where it gets called   | `processor.py` / LightRAG internals   | grep `embedding_func` |
