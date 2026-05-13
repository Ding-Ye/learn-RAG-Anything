# Appendix B · Upstream file map

## Problem

When you go from this teaching repo to the real upstream, you'll find
`HKUDS/RAG-Anything @ 146828f7…` has ~12K lines of Python across 19
files. Without a map it's not obvious where to land first. This
appendix is that map.

## Solution

A single table: every upstream file we care about, what it owns, and
which of our chapters maps to it. Drop this on a side monitor when you
read the upstream code.

## How It Works

The upstream layout (under `raganything/`):

| File                     | LOC   | Owns                                                                 | Our chapter   |
| ------------------------ | ----: | -------------------------------------------------------------------- | ------------- |
| `__init__.py`            |   114 | Public surface; re-exports `RAGAnything`                              | —             |
| `base.py`                |    12 | `DocStatus` enum                                                     | s01 vocabulary |
| `config.py`              |   158 | `RAGAnythingConfig` dataclass + env-var loading                       | s08 surface    |
| `parser.py`              |  2660 | `Parser` base + MinerU / Docling / PaddleOCR adapters + CLI           | s02            |
| `enhanced_markdown.py`   |   534 | Post-parse markdown normalization (links, asset URLs, etc.)          | s02 (deep)    |
| `asset_urls.py`          |   117 | Attach public URLs to image / table assets                            | s02 (deep)    |
| `processor.py`           |  2229 | `ProcessorMixin`: chunking, ingest, entity/relation extraction       | s03, s08       |
| `modalprocessors.py`     |  1607 | Image / Table / Equation / Generic modal processors + `ContextExtractor` | Appendix A     |
| `omml_extractor.py`      |   758 | OMML (Office math) → LaTeX conversion                                | Appendix A     |
| `utils.py`               |   380 | Helpers: `get_processor_for_type`, `get_table_body`, image base64, etc. | s01, s03, Appendix A |
| `query.py`               |   868 | `QueryMixin`: text / multimodal query, cache keys                     | s06            |
| `prompt.py`              |   406 | `PromptRegistry` + English prompt templates                          | s07            |
| `prompts_zh.py`          |   337 | Chinese prompt translations                                          | s07            |
| `prompt_manager.py`      |   156 | Language switch, per-call overrides                                  | s07            |
| `resilience.py`          |   397 | Retry / timeout primitives (bounded retries, exponential backoff)    | (future s09)   |
| `batch.py`               |   428 | `BatchMixin`: process N documents with progress reporting            | (future s10)   |
| `batch_parser.py`        |   470 | Standalone batch-parsing CLI                                          | (future s10)   |
| `callbacks.py`           |   377 | Progress / stage callbacks for long ingests                           | (future)       |
| `raganything.py`         |   644 | Orchestrator class composing the mixins                              | s08            |

Total: ~12.6K LOC of Python.

## What Changed

Nothing in the codebase. This appendix is a reference index — read it
side-by-side with `upstream-readings/sNN-*.go.md`, which has the
deeper per-chapter excerpts.

## Try It

Order to read the upstream if you want a fast tour:

1. `__init__.py` — see what's public.
2. `config.py` — see what knobs RAGAnything exposes.
3. `raganything.py` — see the orchestrator and constructor.
4. `parser.py` — first half (class declarations, format constants).
5. `processor.py` — the `_process_text_chunks` path.
6. `modalprocessors.py` — `ImageModalProcessor.process_modal` start.
7. `query.py` — the `aquery` path.
8. `prompt.py` — the registry mechanism + a couple of templates.

Each of our chapters anchors one of these stops.

## Upstream Source Reading

Two file-level notes worth knowing:

- **`parser.py` is huge** (2.6K LOC) but most of it is platform-specific
  shell-out logic for MinerU / Docling / PaddleOCR. The interesting
  Python lives in the first ~200 lines and in `MineruParser.parse`.
- **`processor.py` is also huge** (2.2K LOC) and most of it is
  multimodal coordination. The teaching-relevant code is
  `_process_text_chunks` and `_generate_cache_key`.

A practical strategy: keep `git grep` open on the upstream tree and
jump by symbol name as you read each `upstream-readings/sNN-*.go.md`
in this repo.
