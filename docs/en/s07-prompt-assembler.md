# s07 · prompt-assembler

## Problem

We have retrieved chunks and the user's question. We do *not* have an
answer yet — that's the LLM's job. But the LLM is only as good as the
prompt we hand it. A bad prompt loses citations, exceeds the context
window silently, or fails to instruct the model to stay within the
retrieved evidence. Concretely:

- How do we tell the LLM "answer ONLY from this context"?
- How do we make citations possible (so the user can verify)?
- How do we fit a budget without dropping the *highest-scored* chunk?
- How do we serve a multilingual product without rewriting the template?

## Solution

A single `AssemblePrompt(question, hits, opts)` that returns:

- `Text` — the prompt string to ship to the LLM,
- `UsedChunks` — the chunks that made it into the prompt,
- `DroppedCount` — how many were truncated for size.

Behavior:

- **Bilingual templates** (`LangEN` / `LangZH`) mirror upstream's
  `prompt.py` + `prompts_zh.py` split.
- **Citation markers** (`[c1] [c2] …`) are toggled by `IncludeIDs`, so
  the LLM can cite by index.
- **Score-aware truncation**: if `MaxChars > 0`, drop the
  *lowest-scored* chunk first and keep retrying until the joined
  context fits. Highest-scored chunks always survive.
- **MinScore floor** cuts noise before truncation runs.

## How It Works

1. Filter hits by `MinScore`, then sort by descending score.
2. While the joined context exceeds `MaxChars`, pop the last (lowest-
   scored) chunk and count it as dropped.
3. Render via `renderTemplate(lang, question, hits, ids)` — two near-
   identical functions that produce the EN/ZH structure with the same
   ordering of context → question.

The assembler depends on **nothing from upstream stages**: it accepts
`[]rag.RetrievedChunk` and emits a string. That's deliberate; you can
swap retrievers without touching the prompt.

## What Changed

- First stage that produces a *string-shaped* artifact (a prompt)
  rather than a Go struct.
- Introduces explicit budget management — earlier chapters had no
  notion of "doesn't fit".

## Try It

```bash
go run ./agents/s07-prompt-assembler
```

Expected (truncated):

```
--- EN prompt ---
You are a careful assistant. Answer the QUESTION using ONLY the CONTEXT below. ...

CONTEXT:
[c1] Retrieval-Augmented Generation grounds an LLM in external documents. (section=Intro)
[c2] Chunks are embedded as dense vectors and stored in a vector index. (section=Pipeline)
...

QUESTION:
What does RAG actually do?

ANSWER:

--- ZH prompt ---
你是一个严谨的助手。请仅依据【参考资料】回答【问题】，...
```

Tests:

```bash
go test ./agents/s07-prompt-assembler
```

## Upstream Source Reading

Upstream stores its prompt library in `raganything/prompt.py` (English)
and `raganything/prompts_zh.py` (Chinese), with a `PromptRegistry`
allowing atomic language switches at runtime. The actual selection of
which template to use for a given content type happens in
`raganything/prompt_manager.py`. See
[`upstream-readings/s07-prompts.go.md`](../../upstream-readings/s07-prompts.go.md)
for the `PromptRegistry.swap` mechanism and the bilingual key tables.
