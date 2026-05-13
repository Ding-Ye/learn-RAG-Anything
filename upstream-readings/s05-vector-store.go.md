# Upstream reading · vector store (s05)

> `.go.md` extension so `go build` ignores this file.

**Source**: `raganything/raganything.py` from
`HKUDS/RAG-Anything @ 146828f7…` (MIT).

The vector store itself is owned by **LightRAG** (a separate package);
RAGAnything is its consumer. From the orchestrator's perspective:

```python
# raganything/raganything.py  (sketch, comments added)

class RAGAnything(...):
    def __init__(self, working_dir, ..., embedding_func=None, ...):
        # LightRAG owns three separate vector storages:
        #   - text_chunks    (chunk embeddings)
        #   - entities       (entity-graph node embeddings)
        #   - relationships  (entity-graph edge embeddings)
        # All three share the same interface.
        self.lightrag = LightRAG(
            working_dir=working_dir,
            embedding_func=embedding_func,
            # ... storage backends pluggable here:
            vector_storage="NanoVectorDBStorage",   # default
            kv_storage="JsonKVStorage",
            graph_storage="NetworkXStorage",
        )
```

The vector-storage backend is selected by **name**, and at least the
following are available in LightRAG: `NanoVectorDBStorage` (in-process),
`MilvusVectorDBStorage`, `ChromaVectorDBStorage`, `FaissVectorDBStorage`,
`QdrantVectorDBStorage`, plus a PostgreSQL/`pgvector` variant.

The interface every backend implements is roughly:

```python
class VectorDBStorage:
    async def upsert(self, data: dict[str, dict]) -> None: ...
    async def query(self, query: str, top_k: int) -> list[dict]: ...
    async def delete_entity(self, entity_name: str) -> None: ...
    async def index_done_callback(self) -> None: ...
    # plus other lifecycle hooks
```

## Key observations

1. **The interface is upsert + query.** Same idea as our `Add` +
   `Search`, just with async/await and dict payloads.
2. **There are *three* vector stores**, not one — chunks vs. KG nodes
   vs. KG edges. Each has its own embeddings. Our teaching pipeline
   only models the chunk store, which is the central one; the graph
   pieces are out of scope for this curriculum.
3. **Backends are selected by string at construction time.** This is a
   plugin pattern; it lets the same code talk to FAISS, Milvus, or a
   tiny in-process index.
4. **The query is `(text, top_k)`, not `(embedding, top_k)`.** Upstream
   embeds the query *inside* the store call. We split that into s06
   (retriever) so the embed step is visible.

## Map to our Go types

| Upstream                                  | This repo                          |
| ----------------------------------------- | ---------------------------------- |
| `VectorDBStorage.upsert({...})`           | `VectorStore.Add(rag.VectorRecord)`|
| `VectorDBStorage.query(text, top_k)`      | `VectorStore.Search(query, k)`     |
| Backend selection by string               | one concrete `MemoryStore`         |
| Three separate stores (chunks/entities/edges) | one store for chunks only      |
| Query embed inside `.query`               | embed in s06 retriever, search here |

## Where to look in upstream

| Concept                       | File                       | Hint                              |
| ----------------------------- | -------------------------- | --------------------------------- |
| Storage backends listed       | `raganything/raganything.py` | grep `vector_storage`           |
| LightRAG storage interfaces   | LightRAG repo (external)   | `lightrag.storage.*`              |
| Where chunks get upserted     | `raganything/processor.py` | grep `text_chunks` / `upsert`     |
