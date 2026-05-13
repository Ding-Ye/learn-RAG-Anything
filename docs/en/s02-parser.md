# s02 · parser

## Problem

A real RAG framework starts from PDFs, DOCX, slides, images. Before
anything else can happen, those bytes need to become a stream of typed
units we can chunk and embed. Upstream calls this layer the "parser",
and delegates to MinerU, Docling, or PaddleOCR depending on the format.
Two questions matter for learners:

1. What does the output of a parser actually *look like*?
2. How does the parser preserve provenance (page, section) so we can
   later answer "where did this come from"?

## Solution

We strip away the messy file-format I/O and focus on the transform
itself: text source → `[]rag.Block`. Our toy markdown parser
(`ParseMarkdown`) handles:

- ATX headings (`#`, `##`, ...): emit a `BlockHeading` and start a new
  section.
- Fenced code (`` ``` ``): emit a `BlockCode` carrying the language tag.
- Paragraphs: any non-marker line accumulates until a blank line, then
  emits as `BlockText`.
- `---` on its own line: page-break marker that increments the `Page`
  counter.

About 100 lines of Go, no dependencies.

## How It Works

`ParseMarkdown` is a single forward scan over `lines := split(src, "\n")`:

1. Keep two pieces of running state: `section` (last heading title)
   and `page` (incremented at every `---`).
2. Maintain a `paragraph []string` buffer. Blank lines / heading lines
   / code-fence lines all *flush* it into a `BlockText`.
3. When we hit a `` ``` `` fence, fast-forward to the closing fence and
   emit one `BlockCode` carrying the inner lines verbatim.
4. Every emit assigns the current `Page`, `Section`, and a monotonic
   `Order` so later stages can recover document order.

Crucially the parser knows nothing about embeddings, retrieval, or
prompts — it just owns the `bytes → []Block` boundary.

## What Changed

- Added `agents/s02-parser/main.go` with `ParseMarkdown`.
- The first stage in our pipeline that actually transforms data; s01
  only declared types.

## Try It

```bash
go run ./agents/s02-parser
```

Expected output (truncated):

```
parsed 6 blocks
  [heading] page=1 section="RAG-Anything (mini)" lang=    text="RAG-Anything (mini)"
  [text]    page=1 section="RAG-Anything (mini)" lang=    text="Retrieval-Augmented Generation lets an LLM answer u…"
  [heading] page=1 section="Pipeline"             lang=    text="Pipeline"
  [text]    page=1 section="Pipeline"             lang=    text="The pipeline has four stages: parse, embed, retri…"
  [code]    page=1 section="Pipeline"             lang=go  text="chunks := store.Search(rag.Embed(query), 5)"
  [heading] page=2 section="Multi-modal"          lang=    text="Multi-modal"
  ...
```

Tests:

```bash
go test ./agents/s02-parser
```

## Upstream Source Reading

Upstream `raganything/parser.py` defines a `Parser` base with format
constants and delegates extraction to MinerU / Docling / PaddleOCR.
The reading at
[`upstream-readings/s02-parser.go.md`](../../upstream-readings/s02-parser.go.md)
walks through that base class, the format-routing constants
(`OFFICE_FORMATS`, `IMAGE_FORMATS`, `TEXT_FORMATS`), and shows how the
upstream output already resembles our `[]rag.Block` shape: a list of
markdown blocks each tagged with `type`, `page_idx`, and
caption/footnote arrays.
