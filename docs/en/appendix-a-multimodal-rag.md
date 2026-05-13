# Appendix A · Multi-modal RAG

## Problem

This curriculum builds a *text-only* pipeline. Real documents are
messier: a single page can mix narrative text, a table, a math
equation, and a figure with caption and footnotes. Upstream
`HKUDS/RAG-Anything` is named for exactly this reason — it tries to be
"all-in-one RAG for multi-modal documents". So: what actually changes
when you add modalities, and why didn't we model it in code?

## Solution

This appendix takes the four common non-text modalities — **image,
table, equation, OCR-rendered glyphs** — and for each one explains:

1. how upstream represents it,
2. what new external system it depends on,
3. where the pipeline forks, and
4. what stays the same as the text-only path.

No new Go code. The goal is to ground the existing curriculum in
reality and tell you exactly which existing chapters would change if
you extended this repo into multi-modal land.

## How It Works

### Image blocks

Upstream `ImageModalProcessor` (`raganything/modalprocessors.py`) is a
small state machine:

1. **Caption.** Send the image to a **vision LLM** (`vision_model_func`)
   along with the `vision_prompt_with_context` template from
   `raganything/prompt.py`. The model returns JSON with
   `detailed_description` and an `entity_info` block.
2. **Embed the caption** — *not* the image. The caption text rides
   the same `embedding_func` as ordinary text chunks, so retrieval is
   uniform.
3. **Index.** The caption becomes a chunk-like record; the image path
   is preserved as metadata so the answer can display the image when
   it cites the chunk.

What would change in our Go code: `BlockKind = BlockImage`, a new
`Captioner` interface, and a tiny dispatch in the chunker.
Everything downstream stays unchanged.

### Table blocks

Tables are *linearized*, not embedded as images. The upstream
`TableModalProcessor` does this in two steps: extract the table body
as cleaned HTML/markdown (`get_table_body` in `raganything/utils.py`),
then optionally ask an LLM to write a natural-language summary using
`TABLE_ANALYSIS_SYSTEM`. Both the summary and the raw body are kept;
embedding goes against the summary (queries are usually phrased as
prose, not as table cells).

### Equation blocks

Equations are stored as **TeX**, never as a rendered image, and the
`EquationModalProcessor` asks the LLM for a textual analysis of what
the equation means (variables, common uses, sample numeric example).
Embedding goes against the analysis, not the raw TeX.

There's a fun edge case here: Office equations come in OMML, not
TeX. `raganything/omml_extractor.py` converts OMML → TeX before the
modal processor runs. This is the kind of fiddly real-world detail
that disappears in a teaching impl.

### OCR'd glyphs (scanned PDFs)

Pure-text blocks pulled out by an OCR engine (PaddleOCR in the
upstream default) are tagged with confidence scores. The pipeline
treats them like normal text but uses confidence to weight the
ranking: low-confidence OCR text gets a penalty before reaching the
vector store. We don't model this in Go but the right place would be
`Block.Meta["confidence"]` + a hook in s05's Search.

## What Changed

Nothing in the existing chapters. This appendix is a map of *future
extensions*. If you wanted to grow this repo into a multimodal one,
the minimum-change recipe is:

| Modality | Add                     | Touches                  |
| -------- | ----------------------- | ------------------------ |
| Image    | `Captioner` interface   | s02, s03, new `s09-image` |
| Table    | `TableLinearizer`       | s02, s03                 |
| Equation | TeX/OMML decoder        | s02, s03                 |
| OCR      | `Confidence` metadata   | s02, s05                 |

Every modality eventually collapses into "a string the embedder can
see", so the vector store, retriever, and prompt assembler do **not**
change. That insight — that multi-modal RAG is really "many ways to
produce text" rather than "many vector spaces" — is the appendix's
single takeaway.

## Try It

This appendix doesn't ship runnable code, but you can browse the
upstream files to see real implementations:

```bash
less /path/to/upstream/raganything/modalprocessors.py
# search for: class ImageModalProcessor, class TableModalProcessor, ...
less /path/to/upstream/raganything/omml_extractor.py
```

## Upstream Source Reading

| Modality | File                              | Class / Function             |
| -------- | --------------------------------- | ---------------------------- |
| Image    | `raganything/modalprocessors.py`  | `ImageModalProcessor`        |
| Table    | `raganything/modalprocessors.py`  | `TableModalProcessor`        |
| Equation | `raganything/modalprocessors.py`  | `EquationModalProcessor`     |
| OMML→TeX | `raganything/omml_extractor.py`   | `extract_omml_to_latex`      |
| Generic  | `raganything/modalprocessors.py`  | `GenericModalProcessor`      |
| Prompts  | `raganything/prompt.py`           | `vision_prompt_with_context`, `TABLE_ANALYSIS_SYSTEM`, etc. |
| Utility  | `raganything/utils.py`            | `get_table_body`, `get_equation_text_and_format` |
