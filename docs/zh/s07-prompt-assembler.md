# s07 · prompt-assembler

## Problem

我们手里已经有检索回来的 chunks 和用户的问题，但还*没有*答案——那
是 LLM 的事。但 LLM 的好坏完全取决于我们递给它的 prompt：糟糕的
prompt 会丢失 citation、悄无声息超出上下文窗口、或者根本没让模型
守在检索到的证据范围内。具体来说：

- 怎么告诉 LLM"只能基于这段上下文回答"？
- 怎么让 citation 成为可能（让用户能验证）？
- 怎么在预算内尽量保住*分数最高*的 chunk？
- 怎么不重写模板就同时支持中英文？

## Solution

一个 `AssemblePrompt(question, hits, opts)` 返回：

- `Text` —— 发给 LLM 的 prompt 字符串；
- `UsedChunks` —— 真正进了 prompt 的 chunks；
- `DroppedCount` —— 因为预算被砍掉了几条。

行为：

- **双语模板**（`LangEN` / `LangZH`），对应上游
  `prompt.py` + `prompts_zh.py` 的拆分。
- **Citation 标记**（`[c1] [c2] …`），由 `IncludeIDs` 开关，让
  LLM 在答案里按编号引用。
- **基于分数的截断**：如果 `MaxChars > 0`，从*分数最低*的 chunk
  开始砍，直到 context 字符数符合预算；分数最高的 chunk 永远不会
  被砍。
- **MinScore 阈值** 在截断之前先剔除噪声。

## How It Works

1. 按 `MinScore` 过滤，然后按分数降序排。
2. 只要 context 字符数 > `MaxChars`，弹出最后一条（分数最低）并
   计入 dropped 计数。
3. 调用 `renderTemplate(lang, question, hits, ids)`——两份几乎对称
   的函数，分别产出 EN/ZH 结构：context 在前、question 在后。

assembler 不依赖任何上游阶段：输入是 `[]rag.RetrievedChunk`，输出
是字符串。这是刻意的；换 retriever 不需要动 prompt。

## What Changed

- 第一阶段产物从 Go 结构变成了*字符串*（prompt）。
- 引入显式预算管理——前面几章对"装不下"没有概念。

## Try It

```bash
go run ./agents/s07-prompt-assembler
```

预期输出（截断）：

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

测试：

```bash
go test ./agents/s07-prompt-assembler
```

## Upstream Source Reading

上游把 prompt 库放在 `raganything/prompt.py`（英文）和
`raganything/prompts_zh.py`（中文），用 `PromptRegistry` 提供
运行时的原子语言切换。具体到某种内容类型选哪个模板，是在
`raganything/prompt_manager.py` 里挑的。导读见
[`upstream-readings/s07-prompts.go.md`](../../upstream-readings/s07-prompts.go.md)。
