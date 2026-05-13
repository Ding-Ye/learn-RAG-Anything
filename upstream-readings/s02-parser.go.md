# Upstream reading · parser (s02)

> `.go.md` extension so `go build` ignores this file.

**Source**: `raganything/parser.py` from
`HKUDS/RAG-Anything @ 146828f7…` (MIT).

## The shape of the base class

```python
# raganything/parser.py  (excerpt, comments added)

class Parser:
    """Base class for document parsing utilities."""

    # Format routing tables — the parser dispatches by extension.
    OFFICE_FORMATS = {".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx"}
    IMAGE_FORMATS  = {".png", ".jpeg", ".jpg", ".bmp", ".tiff",
                      ".tif", ".gif", ".webp"}
    TEXT_FORMATS   = {".txt", ".md"}
```

A concrete parser (e.g. `MineruParser`) overrides `parse()` to shell
out to `mineru` and read back the resulting markdown + asset table.
Inside that table, each item already looks roughly like:

```python
{
    "type": "text"       # or "image" | "table" | "equation" | "code"
    "text": "...",
    "page_idx": 3,
    "captions": [...],
    "footnotes": [...],
    "img_path": "media/figure-1.png",   # only for images
    "table_body": "<html>...</html>",   # only for tables
}
```

That dict-of-dicts is exactly what we have as `rag.Block` — minus the
fields we don't need for a teaching repo.

## Key observations

1. **Format routing is by extension, not by content sniffing.** Cheap
   and predictable; we don't even bother with this since we only consume
   markdown.
2. **A page index travels with every block.** This is the reason we
   put `Page int` on `rag.Block` in s01.
3. **Captions and footnotes are arrays, not strings.** Multi-modal
   parsers attach several captions to one image; we keep this in mind
   for the appendix even though our toy parser doesn't emit images.
4. **The parser is a transformation, not a session.** It is called once
   per document and returns data; no streaming, no state held between
   documents. Our Go function `ParseMarkdown` mirrors that.

## Map to our Go types

| Upstream                              | This repo                                     |
| ------------------------------------- | --------------------------------------------- |
| dict with `"type"` + `"text"`         | `rag.Block.Kind` + `rag.Block.Text`           |
| dict with `"page_idx"`                | `rag.Block.Page`                              |
| MinerU output: list of these dicts    | `[]rag.Block` returned by `ParseMarkdown`     |
| `register_parser(name, cls)`          | not needed yet — a single parser, no plugin   |
| Heading detected by MinerU heuristics | `# ` / `## ` prefix in `ParseMarkdown`        |
| Code blocks tagged with language      | `Block.Lang` after stripping the ` ``` lang ` |

## Where to look in upstream

| Concept             | File                       | Hint                                |
| ------------------- | -------------------------- | ----------------------------------- |
| `Parser` base       | `raganything/parser.py`    | search "class Parser:"              |
| MinerU subclass     | `raganything/parser.py`    | search "class MineruParser"         |
| Format routing      | `raganything/parser.py`    | `OFFICE_FORMATS`, `IMAGE_FORMATS`   |
| Plugin registry     | `raganything/parser.py`    | `register_parser`                   |
| Block-list consumer | `raganything/processor.py` | search "content_list" handling      |
