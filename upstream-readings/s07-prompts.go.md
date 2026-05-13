# Upstream reading · prompt assembler (s07)

> `.go.md` extension so `go build` ignores this file.

**Sources**:

- `raganything/prompt.py` (English)
- `raganything/prompts_zh.py` (Chinese)
- `raganything/prompt_manager.py` (selection logic)

All from `HKUDS/RAG-Anything @ 146828f7…` (MIT).

## The `PromptRegistry` swap trick

Upstream has a clever data structure for runtime language switching:

```python
# raganything/prompt.py  (excerpt, comments added)

class PromptRegistry:
    """Stable prompt container with atomic snapshot swapping."""

    def __init__(self) -> None:
        self._data: dict[str, Any] = {}

    def swap(self, prompts: dict[str, Any]) -> None:
        """Atomically replace the active prompt snapshot."""
        self._data = dict(prompts)

    # __getitem__, __contains__, etc. — behaves like a dict.

PROMPTS = PromptRegistry()

PROMPTS["IMAGE_ANALYSIS_SYSTEM"] = (
    "You are an expert image analyst. ..."
)
PROMPTS["TABLE_ANALYSIS_SYSTEM"] = (
    "You are an expert data analyst. ..."
)
# ... vision_prompt, vision_prompt_with_context, ...
```

Then `prompts_zh.py` builds a parallel dict and provides a tiny helper
that swaps into the registry:

```python
# raganything/prompts_zh.py  (idea)
ZH_PROMPTS = {
    "IMAGE_ANALYSIS_SYSTEM": "你是一名专业的图像分析师……",
    ...
}

def use_chinese_prompts():
    PROMPTS.swap(ZH_PROMPTS)
```

`prompt_manager.py` exposes a small surface for users to enable
Chinese, query a specific template, or override per-call.

## Why this matters

1. **Readers and writers don't race.** A consumer reads via
   `PROMPTS["KEY"]`; a switch replaces `_data` in one assignment. There
   is no per-key locking, no half-translated state.
2. **Prompts are *content*, not code.** Switching languages is data,
   not class hierarchies. That's exactly the spirit of our
   `renderTemplate(lang, ...)` switch.
3. **Multiple prompt families coexist.** Upstream prompts cover image,
   table, equation, and generic content. The same registry holds them
   all keyed by name.

## Map to our Go types

| Upstream                                | This repo                                |
| --------------------------------------- | ---------------------------------------- |
| `PROMPTS["IMAGE_ANALYSIS_SYSTEM"]`      | a string constant inside `renderTemplate`|
| `PROMPTS.swap(ZH_PROMPTS)`              | `AssembleOptions.Lang`                   |
| `prompt_manager.use_chinese_prompts()`  | callers pass `LangZH`                    |
| `vision_prompt_with_context` template   | not implemented (appendix material)      |
| Atomic dict swap (thread safety)        | no concurrency yet — single-call API     |

## Where to look in upstream

| Concept              | File                              | Hint                                    |
| -------------------- | --------------------------------- | --------------------------------------- |
| `PromptRegistry`     | `raganything/prompt.py`           | top of file                             |
| EN prompts           | `raganything/prompt.py`           | grep `PROMPTS[`                         |
| ZH prompts           | `raganything/prompts_zh.py`       | full file                               |
| Selection / swap     | `raganything/prompt_manager.py`   | `use_chinese_prompts` and friends       |
| Vision prompt        | `raganything/prompt.py`           | `vision_prompt_with_context`            |
