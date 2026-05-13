# s08 · pipeline

## Problem

s02..s07 各演示了*一个*机制。下一步自然是把它们串起来：从前端塞
一份 markdown 进去，在后端提问，看检索如何引导答案。本章就是把
抽象变具体的地方。

它也逼出一个设计决定：编排者是否要知道每个阶段的细节？答案在真实
系统（以及本章）里都是否定的——pipeline 之所以这么小，就因为每个
阶段都把自己的逻辑藏在接口后面。

## Solution

`Pipeline` 有三个字段：

```go
type Pipeline struct {
    Embedder Embedder
    Store    VectorStore
    LLM      LLM
}
```

两个方法：

- `Ingest(docID, source, markdown) int` —— parse → chunk → embed → store。
- `Ask(question, k) (answer, hits, prompt)` —— embed → retrieve →
  assemble → complete。

为了让本章自包含，前几章的实现被压缩成同一个 `package main` 里的
私有 helper，每段都是其原章节的"忠实瘦身版"并附注释指回。任何一段
都能换成对应章节的完整实现或者生产后端，因为接口
(`Embedder`、`VectorStore`、`LLM`) 都极小。

## How It Works

`main` 用一段简短但贴近现实的 markdown 喂给 pipeline：

1. `NewPipeline()` 接好 64 维哈希 embedder、内存 store、`MockLLM`。
   MockLLM 不打外网——它从 prompt 里抠出 `CONTEXT:`，回显第一个
   chunk，让你一眼看到答案确实来自检索。
2. `Ingest("rag-mini", "rag-mini.md", corpus)` 跑一遍四阶段建索，
   并报告产生了多少 chunk。
3. `Ask(q, 3)` 对几条问题各跑一次：打印 top-3 chunk id、prompt
   长度，以及那条引用 `[c1]` 的合成回答。

数据流就是上游文档画的那张图：

```
markdown → parser → []Block → chunker → []Chunk
                                     ↓
                                  embedder
                                     ↓
                                  VectorStore
question → embedder → Search → []RetrievedChunk → prompt → LLM → answer
```

## What Changed

- 第一章把前几章机制编排起来。
- 引入 `LLM` 接口和 `MockLLM`，pipeline 不需要任何 API key 就能跑。
- 通过把每个阶段都接成接口，演示了"可替换性"这件事。

## Try It

```bash
go run ./agents/s08-pipeline
```

预期输出（截断）：

```
ingested 12 chunks into the store

Q: What do embeddings do?
  hit[1] score=+0.452 id=rag-mini-004-fb5d5812
  hit[2] score=+0.126 id=rag-mini-003-e12ca782
  hit[3] score=+0.121 id=rag-mini-010-29d646e7

  prompt-len=373 chars
  A: (mock answer) Based on the top-ranked chunk, [c1] Embeddings ...
```

测试：

```bash
go test ./agents/s08-pipeline
```

想看一次完整查询的端到端走读（含对上游文件的引用），可参考
[`docs/zh/s_full-integration.md`](./s_full-integration.md)。

## Upstream Source Reading

上游的编排逻辑在 `raganything/raganything.py`：`process_document_complete`
负责 ingest，`aquery` / `aquery_with_multimodal` 负责提问；
它们坐在 `ProcessorMixin`、`QueryMixin`、`BatchMixin` 几个 mixin 栈
的顶部。导读见
[`upstream-readings/s08-pipeline.go.md`](../../upstream-readings/s08-pipeline.go.md)。
Python 的"按关注点分 mixin"，对应我们这里"在 struct 里组合接口"。
