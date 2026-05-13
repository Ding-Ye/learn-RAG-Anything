# s01 · block-model

## Problem

任何一个 RAG 框架最终都要回答一个看似不起眼的问题：
"在内部，一段内容到底*是什么*？"PDF 的一页里，可能混着正文、表格、
公式、图注、代码块。如果你的流水线把它们一律当作 `string`，那基本就
已经输了 —— Chunker 没办法把表格保持成一个整体；Embedder 没办法对
图片单独走一条 captioning 路径；Retriever 也没办法告诉用户"这条
答案来自第 4 页的表格"。所以在动手解析之前，先得把类型词表立起来。

## Solution

s01 在 `rag/types.go` 里把这套词表定下来，并用一个手写的样例文档把
它展示出来。Go 类型有：

- `BlockKind`：枚举 `text | heading | code | image | table | equation`。
- `Block`：一个带类型的最小单元，有 `Order`、`Page`、`Section`、`Meta`。
- `Document`：`ID`、`Source`、`[]Block`。
- （供后续章节用的占位）`Chunk`、`Embedding`、`VectorRecord`、`RetrievedChunk`。

本章 `main.go` 手工构造一个 3-block 的样例文档，先用文本摘要打印
一遍，再打印一份完整 JSON，让你在算法登场前就先看清楚数据形态。

## How It Works

1. `buildSampleDoc` 返回一个 `rag.Document` 字面量，里面有三个 block：
   heading、text、code。
2. `summarize` 遍历这些 block，每条打印一行并按 `BlockKind` 打标签。
3. `main` 先打印 summary，再用 `json.Encoder` 输出格式化 JSON，
   让 wire 上的数据形态肉眼可见。

`rag/` 包刻意只放类型，不放任何行为。后面每一章都把自己的具体行为
写在自己的目录里。

## What Changed

这是第一章，没有"上一章"。这一节存在的意义就是：在任何 parser、
embedder、retriever 出现之前，先把词表固定下来。

## Try It

```bash
go run ./agents/s01-block-model
```

预期输出（截断）：

```
Document doc-001 (source=intro.md, 3 blocks)
  [heading] order=0 page=1 section="intro" text="RAG in one paragraph"
  [text] order=1 page=1 section="intro" text="Retrieval-Augmented Generation lets an LLM answer using d…"
  [code] order=2 page=1 section="intro" text="rag.Embed(query) -> top-k chunks -> prompt(LLM)"

Raw JSON:
{
  "ID": "doc-001",
  ...
}
```

测试：

```bash
go test ./agents/s01-block-model
```

## Upstream Source Reading

上游把"modality"这件事写在
`raganything/modalprocessors.py`：`ImageModalProcessor`、
`TableModalProcessor`、`EquationModalProcessor`、`GenericModalProcessor`
各管一类。导读节选见
[`upstream-readings/s01-blocks.go.md`](../../upstream-readings/s01-blocks.go.md)，
里面解释了 Python 的 dataclass + `get_processor_for_type` 查表，
是怎么承担"我们这边由 `BlockKind` 枚举承担"的角色的。
