# Upstream reading · chunker (s03)

> `.go.md` extension so `go build` ignores this file.

**Source**: `raganything/processor.py` from
`HKUDS/RAG-Anything @ 146828f7…` (MIT).

Chunking in upstream lives in `ProcessorMixin._process_text_chunks`
inside `processor.py`. It is more complex than ours because it has to
coordinate with LightRAG's storage (chunk hashes, status tracking,
async batching). The teaching nugget is the size policy:

```python
# raganything/processor.py  (sketch, comments added)

class ProcessorMixin:
    async def _process_text_chunks(self, content_list, doc_id):
        # 1) Concatenate text blocks into a single string per document,
        #    preserving inline asset markers (images/tables/equations).
        # 2) Delegate the *actual* chunk splitting to LightRAG, which
        #    uses the tokenizer of the embedding model:
        #
        #        from lightrag.utils import compute_mdhash_id
        #        chunks = self.lightrag.chunk_text(text, ...)
        #
        # 3) For each chunk produce a dict that ends up looking like:
        #
        #        {
        #          "content": "...chunk text...",
        #          "tokens":  178,
        #          "chunk_order_index": 4,
        #          "full_doc_id": doc_id,
        #          "_id": compute_mdhash_id(content, prefix="chunk-"),
        #        }
        #
        # 4) Persist to LightRAG.text_chunks storage.
```

## Key observations

1. **Tokenizer-based budget, not char-based.** Upstream uses the
   embedding-model tokenizer so a chunk's size measures what the
   embedder will actually consume. Our `TargetChars` is a teaching
   stand-in.
2. **Stable IDs.** `compute_mdhash_id(content, prefix="chunk-")` is a
   content-derived ID. We use the same trick (a SHA-1 prefix of
   `docID|text`) so chunks are addressable and de-duplicatable.
3. **Chunk order is metadata.** `chunk_order_index` lets downstream
   stitch chunks back into document order. We bake the same idea into
   `Meta["index"]` plus the alphabetical sort of our IDs.
4. **Assets are inline.** Upstream keeps "image marker" placeholders
   inside the text so chunks can still reference the asset table when
   re-rendered. Our toy version leaves this for the appendix.

## Map to our Go types

| Upstream                                       | This repo                          |
| ---------------------------------------------- | ---------------------------------- |
| `lightrag.chunk_text(text, ...)`               | `Chunk()` in `s03-chunker/main.go` |
| `tokens` (counted via embedder's tokenizer)    | `ChunkConfig.TargetChars` (chars)  |
| `chunk_order_index`                            | `Meta["index"]`                    |
| `compute_mdhash_id(content, prefix="chunk-")`  | `makeChunkID(docID, idx, text)`    |
| `full_doc_id` linking chunk to document        | `Chunk.DocID`                      |
| Inline image/table markers                     | (skipped — appendix material)       |

## Where to look in upstream

| Concept                | File                          | Hint                          |
| ---------------------- | ----------------------------- | ----------------------------- |
| `ProcessorMixin`       | `raganything/processor.py`    | search "class ProcessorMixin" |
| chunk persistence      | `raganything/processor.py`    | grep `chunk_order_index`      |
| `compute_mdhash_id`    | `lightrag.utils`              | imported at top of `processor.py` |
