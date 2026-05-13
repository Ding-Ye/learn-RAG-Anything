# s06 · retriever

## Problem

用户用自然语言提问，向量库说的是向量和 chunk 的语言。这之间的桥
就是 retriever。真实系统里，retriever 也是各种前/后处理层的归宿：
query 重写、混合检索（BM25 + 向量）、metadata 过滤、reranker、MMR
多样性…… 我们把*最小*的 retriever 单独拎出来，让接口看清楚。

## Solution

一个小小的 `Retriever`，持有 `Embedder` 和 `VectorStore`。
`Retrieve(question, opts)`：

1. 用*与建索时同一个* embedder 把问题向量化（否则分数没有意义），
2. 向 store 请求 top-K 近邻（稍微 overfetch 一点，给过滤留余量），
3. 应用可选的 `SectionFilter`、`MetaFilter`、`MinScore`，
4. 返回不超过 K 条按分数降序的 `RetrievedChunk`。

## How It Works

```go
type RetrieveOptions struct {
    K            int
    SectionFilter string
    MinScore     float32
    MetaFilter   map[string]string
}
```

依赖的两个东西 (`Embedder`、`VectorStore`) 都是接口而非具体类型：
生产里你会传 OpenAI embedder + pgvector；测试里传 s04 的 fake +
s05 的 memory store。retriever 本身不需要改。

Overfetch 很关键：如果 `K=5` 且加了 section 过滤，top-5 可能被几乎
全部过滤掉。我们抓 `3*K`（最少 10），让过滤有预算。真实系统也是
这么干的，只是数字大得多。

## What Changed

- 第一次出现"组合上几章"的章节：从 `rag/` 没有引入新类型，只是
  重新声明了它需要的接口。
- 把元数据过滤的语义放在 retrieval 边界上，store 完全不知道
  metadata 的细节。

## Try It

```bash
go run ./agents/s06-retriever
```

预期输出（截断）：

```
Q: "How does retrieval work?"
  #1 score=+0.291 section="Pipeline" text="Retrieval finds top-k similar chunks"
  #2 score=+0.188 section="Pipeline" text="RAG combines retrieval with generation"
  ...

Q: "blocks" (section=Parsing)
  #1 score=+0.502 text="Markdown headings become heading blocks"
  #2 score=+0.000 text="Page numbers travel with each block"
```

测试：

```bash
go test ./agents/s06-retriever
```

## Upstream Source Reading

上游的 retrieval 在 `raganything/query.py`，但大部分活其实是 LightRAG
做的。`query.py` 里真正有意思的是*多模检索*：除了文本 chunk，还能从
实体图谱里捞，并且通过 `lightrag.QueryParam` 提供
`naive` / `local` / `global` / `hybrid` / `mix` 几种模式。导读见
[`upstream-readings/s06-retriever.go.md`](../../upstream-readings/s06-retriever.go.md)。
