# Upstream reading · block model (s01)

> File extension `.go.md` so `go build` ignores it — these are commented
> Python excerpts for educational reference, not Go source.

**Source**: `raganything/modalprocessors.py` from
`HKUDS/RAG-Anything @ 146828f7…` (MIT).

The upstream framework expresses "modality" with a small family of
per-type processor classes, all sharing a base. The relevant shape is:

```python
# raganything/modalprocessors.py  (excerpt, comments added)

@dataclass
class ContextConfig:
    """Configuration for context extraction."""
    context_window: int = 1
    context_mode: str = "page"      # "page" | "chunk" | "token"
    max_context_tokens: int = 2000
    include_headers: bool = True
    include_captions: bool = True
    filter_content_types: List[str] = None  # e.g. ["text"]

# ... and further down:

class ImageModalProcessor(...):    # handles image blocks
    ...
class TableModalProcessor(...):    # handles table blocks
    ...
class EquationModalProcessor(...): # handles equation blocks
    ...
class GenericModalProcessor(...):  # fallback
    ...
```

A lookup helper (in `raganything/utils.py`) maps a content-type string
to the matching processor:

```python
def get_processor_for_type(content_type: str):
    # returns one of the *ModalProcessor classes above
    ...
```

## What this teaches us about our Go design

1. **The modality label is a first-class identifier**, not a comment.
   Upstream uses Python string keys ("image", "table", "equation",
   "text"). We use `rag.BlockKind` constants — same idea, type-checked.
2. **Per-modality behavior lives outside the data type.** Upstream
   keeps `Block`-shaped data thin and routes behavior through processor
   classes. Our `rag/types.go` does the same: `Block` has no methods;
   behavior shows up chapter by chapter.
3. **Context is part of every modality.** `ContextConfig` is configured
   once and consulted by every processor; our `Block.Section` /
   `Block.Page` fields play the same provenance role, kept on the
   record itself instead of in a side table.

## Where to look in upstream

| Concept            | Upstream file                                | Lines (approx) |
| ------------------ | -------------------------------------------- | -------------- |
| `ContextConfig`    | `raganything/modalprocessors.py`             | 39–53          |
| Modality classes   | `raganything/modalprocessors.py`             | search "ModalProcessor" |
| Routing helper     | `raganything/utils.py` · `get_processor_for_type` | search        |
| Prompt templates   | `raganything/prompt.py`                      | 67–110         |

## Map to our Go types

| Upstream                   | This repo                       |
| -------------------------- | ------------------------------- |
| string `"text"` / `"image"`| `rag.BlockKind` (`text` / `image`) |
| `dataclass` block record   | `rag.Block`                     |
| `get_processor_for_type`   | switch on `BlockKind` per chapter |
| `ContextConfig.context_mode` | `rag.Block.Section` + later `Chunk.Pages` |
