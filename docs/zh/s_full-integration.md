# s_full-integration · 一次查询的端到端走读

## Problem

每章只讲一个阶段，对学习来说是正确的形状，但章节之间的*接缝*
容易被忽略。本文跟着一条 query —— `"What do embeddings do?"` ——
从 markdown 文本一路走到 LLM 完成，给每一阶段的状态命名，并指出它
在哪个文件里。

## Solution

线性走读。没有新代码：我们只是叙述 `go run ./agents/s08-pipeline`
的过程，解释每一章贡献了什么，并指出上游 Python 源码里对应的位置。

## How It Works

### Step 0 · 输入

一段约 600 字符的 markdown（定义在 `agents/s08-pipeline/main.go` 的
`const corpus`），讲 RAG 的基本概念：三个简短标题、若干段落、一个
分页符 `---`。

### Step 1 · parse（s02）

`parseMarkdown(corpus)` 按行扫描，输出 `[]rag.Block`：

```
[heading] "Retrieval-Augmented Generation"
[text]    "RAG combines retrieval with generation: ..."
[heading] "Pipeline"
[text]    "Pipelines typically have four stages: ..."
[heading] "Embeddings"
[text]    "Embeddings map text to dense vectors: ..."
[heading] "Vector stores"        (page=2，分页符之后)
[text]    "A vector store keeps (id, embedding, metadata) rows. ..."
[heading] "Prompt assembly"
[text]    "Once retrieval returns chunks, ..."
```

每个 block 带 `Section`（最近一次标题）和 `Page`。

### Step 2 · chunk（s03）

`chunkDoc(doc, {TargetChars:220, OverlapChars:30})` 遍历 block：

- 标题成为独立的小 chunk：`Retrieval-Augmented Generation`、
  `Pipeline`、`Embeddings`……
- 文本 block 按滑窗拼接，相邻 chunk 共享 30 个字符的 overlap。

对我们的 corpus 这步会产 12 个 chunk，每个 chunk 含：

- `ID` —— `rag-mini-000-fb5d…` 风格，内容哈希。
- `DocID` —— `rag-mini`。
- `Pages` —— `[1]` 或 `[2]`（在 `---` 之后）。
- `Section` —— 对应标题。
- `Text` —— 至多 ~220 字符。

### Step 3 · embed（s04）

对每个 chunk，`Pipeline.Ingest` 调
`Embedder.Embed(chunk.Text)` 得到 64 维单位长度 `rag.Embedding`。
`FakeEmbedder` 用 SHA-256 散射出确定性向量 —— 共享 token 多的两段
落在余弦空间附近。

### Step 4 · store（s05）

每个 `(chunkID, embedding, chunk)` 三元组打包成 `rag.VectorRecord`，
追加到 `MemoryStore`。建索完成后 `p.Store.Len() == 12`。

### Step 5 · ask 阶段：把 query 也 embed 一次（同样是 s04）

`Pipeline.Ask("What do embeddings do?", 3)` 调
`p.Embedder.Embed(question)`，得到 64 维 query 向量。*关键是建索和
查询用的是同一个 embedder* —— 换 embedder 算出来的分数就没意义。

### Step 6 · retrieve（s05 + s06）

`p.Store.Search(qvec, 3)` 线性扫 12 条记录，
计算 `dot(qvec, rec.Embedding)`（两边都是单位向量，所以余弦退化为
点积），按分数降序取 top 3。

对我们这条 query，top-1 是从 "Embeddings" 标题派生出来的小 chunk
—— 用户问 embedding 时，这正是想要的。

### Step 7 · 组装 prompt（s07）

`renderPrompt(question, hits)` 输出：

```
You are a careful assistant. Answer the QUESTION using ONLY the CONTEXT.

CONTEXT:
[c1] Embeddings
[c2] （次优 chunk 文本）
[c3] （第三 chunk 文本）

QUESTION:
What do embeddings do?

ANSWER:
```

`[c1] [c2] [c3]` 让 LLM 可以按编号引用。

### Step 8 · LLM 完成（s08）

`MockLLM.Complete(prompt)` 从 prompt 里抠出 `CONTEXT:` 块，回显
`[c1]` 作为合成答案。生产环境会把它换成
OpenAI / Anthropic / 本地模型客户端；接口 `Complete(prompt) string`
保持不变。

## What Changed

本文不引入新代码，只是把课程的概念图钉死：

| 阶段      | 章节     | 输出类型                   |
| --------- | -------- | -------------------------- |
| Parse     | s02      | `[]rag.Block`              |
| Chunk     | s03      | `[]rag.Chunk`              |
| Embed     | s04      | 每个 chunk 一条 `rag.Embedding` |
| Store     | s05      | `rag.VectorRecord` 行       |
| Retrieve  | s06      | `[]rag.RetrievedChunk`     |
| Assemble  | s07      | 一份 prompt 字符串          |
| Generate  | s08      | 答案字符串                  |

## Try It

```bash
go run ./agents/s08-pipeline
```

修改 `agents/s08-pipeline/main.go` 里的 corpus 或 chunk 配置，看打印
出来的 chunk ID 怎么变。

## Upstream Source Reading

同一条 query 在上游会按这个顺序经过：

| 阶段      | 上游文件                                          |
| --------- | ------------------------------------------------- |
| Parse     | `raganything/parser.py` (MinerU/Docling)          |
| Chunk     | `raganything/processor.py` (`_process_text_chunks`) |
| Embed     | 注入的 `embedding_func`，由 LightRAG 调用          |
| Store     | `lightrag.storage.*VectorDBStorage`                |
| Retrieve  | `raganything/query.py` + LightRAG `aquery`         |
| Assemble  | `raganything/prompt.py` + `prompt_manager.py`      |
| Generate  | 注入的 `llm_model_func`                            |

配套附录 `appendix-b-upstream-map.md` 列了我们关心的所有上游文件
以及它们各自负责什么。
