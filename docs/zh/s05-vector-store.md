# s05 · vector-store

## Problem

有了 embedding 之后，得有地方把它们*存下来*，并且能问"和这条 query
向量最相近的前 k 个 chunk 是哪些"。生产里这件事是 FAISS / pgvector /
Qdrant / Pinecone / Milvus；底层存储千差万别，但接口是相同的两件事：
`Add(record)` 和 `Search(query, k) -> top-k`。把接口和存储技术混为
一谈是常见误区，本章就是为了把它们分开。

## Solution

定义 `VectorStore` 接口，并提供一个最朴素的实现 `MemoryStore`：
内存切片上的线性扫描。两者加起来不到半屏，正是想强调的事——接口
极小，存储可换。

```go
type VectorStore interface {
    Add(rec rag.VectorRecord)
    Search(query rag.Embedding, k int) []rag.RetrievedChunk
    Len() int
}
```

Search 用余弦相似度——由于 s04 输出的是单位向量，余弦退化为点积。

## How It Works

1. `Add(rec)` 把 `rag.VectorRecord`（chunk id、embedding、chunk
   本体）追加到切片。
2. `Search(query, k)` 用 `dot(query, rec.Embedding)` 给每条打分，
   按分数降序，取前 k。
3. 本章自带一个迷你 embedder（FNV 哈希 + 8 次牛顿迭代开方），让 demo
   不依赖 s04。真实使用时直接传入 s04 的 Embedder 即可。

对教学仓库而言，线性扫描就是正解：

- 复杂度 `O(n*dim)`，行为透明、代码极少；
- 没有 bucket、聚类、ANN 近似；
- 真要切到 FAISS，是*换实现*，不是重写流水线。

## What Changed

- 第一次生产 `[]rag.RetrievedChunk` —— s06、s07 都会消费这个类型。
- 第一次在存储层立接口边界：后面所有章节接 `VectorStore`，不耦合
  具体的 `MemoryStore`。

## Try It

```bash
go run ./agents/s05-vector-store
```

预期：

```
indexed 5 chunks (dim=32)

query: "retrieval and generation"
  #1 score=+0.612 id=a text="RAG combines retrieval with text generation"
  #2 score=+0.236 id=e text="Retrieval picks top-k chunks by similarity"
  ...
```

测试：

```bash
go test ./agents/s05-vector-store
```

## Upstream Source Reading

上游的向量库其实*不在* `raganything/` 里——是依赖 LightRAG 提供的
`VectorDBStorage` 家族（chunks / entities / relations 各一份）。
RAGAnything 在 `raganything.py` 的 `__init__` 里把它们连接起来。
导读见
[`upstream-readings/s05-vector-store.go.md`](../../upstream-readings/s05-vector-store.go.md)，
有上游接口与我们这个小接口的对照。
