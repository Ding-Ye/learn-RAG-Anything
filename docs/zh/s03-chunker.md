# s03 · chunker

## Problem

Embedder 都有输入长度上限（常见是 512 或 8K token）。把一整篇解析
后的文档塞进一个向量？既塞不下，又会失去"是哪一段回答了用户提问"
的可定位能力。所以必须 *chunk*：把 Document 的 Blocks 切成若干段
带重叠、可单独 embedding 的小块，同时保留足够元数据让检索结果能
回溯到原文档、章节、页码。

## Solution

`Chunk(doc, cfg)` 遍历 Block 流，把文本累积到一个缓冲区，超过
`cfg.TargetChars` 就 flush 出一个 `rag.Chunk`。两条策略让结果更
适合检索：

- **标题和代码不进缓冲区。** 标题单独成一个小 chunk，便于直接命中；
  代码 block 保持完整，绝不在函数中间切一刀。
- **Overlap（重叠）。** 每次 flush 都用末尾 `cfg.OverlapChars`
  个字符作为下一个 chunk 的开头，让横跨边界的查询也有机会命中
  同时包含两侧上下文的 chunk。

每个 chunk 都带上 `DocID`、`Section`、`Pages` 出处信息。

## How It Works

1. `flush()` 是核心：把当前缓冲区作为 text，构造一个 `rag.Chunk`
   （id 来自 `sha1(docID|text)`），记下涉及的页码，然后把末尾
   overlap 部分塞回缓冲区，作为下一个 chunk 的种子。
2. 主循环分两条路径：
   - 标题 / 代码：先 flush，然后把这个 block 作为独立 chunk 发出。
   - 文本：追加到缓冲区；如果单个 block 比 target 还大，就在
     block 内部按字符切，保证 chunk 不会无限膨胀。
3. `dedupPages` 把页码去重（跨 2-3 页的 chunk 不该出现两个 2）。

Chunker 不知道 embedding、retrieval 和 LLM 的存在，它只负责
`(Document, ChunkConfig) → []Chunk`。

## What Changed

- 首次出现 `rag.Chunk` 的真实生产者。
- 引入 `ChunkConfig`，让策略外显且可调，避免在函数里写死。

## Try It

```bash
go run ./agents/s03-chunker
```

预期输出（截断）：

```
produced 7 chunks (target=150, overlap=30)
  demo-000-17e341b9 pages=[1] section="Intro"  len=5    text="Intro"
  demo-001-96c598c0 pages=[1] section="Intro"  len=150  text="RAG combines retrieval ..."
  demo-002-d076f332 pages=[1 2] section="Intro" len=151 text="...generation. RAG combines..."
  ...
```

注意标题 "Intro" 自成一个小 chunk，代码 chunk `rag.Embed(query)`
保持完整。

测试：

```bash
go test ./agents/s03-chunker
```

## Upstream Source Reading

上游的分块逻辑写在 `raganything/processor.py`，作为
`process_document_complete` 大流程的一部分被调用。它用 LightRAG
的 tokenizer 控制大小，而不是像我们这里用字符数；chunk id 用
`compute_mdhash_id` 计算。导读见
[`upstream-readings/s03-chunker.go.md`](../../upstream-readings/s03-chunker.go.md)。
