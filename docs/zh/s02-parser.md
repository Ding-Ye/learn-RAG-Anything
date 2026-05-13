# s02 · parser

## Problem

真实的 RAG 框架起点是 PDF、DOCX、幻灯片、图片。要让后面的环节
能跑起来，先得把这些字节流变成一串带类型的最小单元。上游把这一层叫
"parser"，并根据格式分发到 MinerU、Docling 或 PaddleOCR。
学习者真正关心的是两件事：

1. parser 的输出到底*长什么样*？
2. parser 怎么把出处信息（页码、章节）保留下来，让我们后面还能回答
   "这条内容来自哪里"？

## Solution

本章把脏活累活——文件 I/O 和各种格式——通通甩掉，只聚焦在
"文本源 → `[]rag.Block`"这一变换。我们用一个玩具级 markdown
parser (`ParseMarkdown`) 处理这些情况：

- ATX 标题 (`#`, `##` …)：发一个 `BlockHeading`，并开启一个新 section。
- 围栏代码 (`` ``` ``)：发一个 `BlockCode`，同时带上语言标签。
- 段落：除上述标记外的非空行累积到下一处空行时打包成 `BlockText`。
- 独占一行的 `---`：作为分页标记，让 `Page` 计数器 +1。

整段大约 100 行 Go，不依赖任何第三方库。

## How It Works

`ParseMarkdown` 就是一次顺序扫描 `lines := split(src, "\n")`：

1. 维护两个状态：`section`（最近一次见到的标题）和 `page`（遇到 `---`
   就 +1）。
2. 维护一个 `paragraph []string` 缓冲区。遇到空行 / 标题 / 代码围栏
   时，把它*flush*成一个 `BlockText`。
3. 遇到 `` ``` `` 围栏就快进到匹配的结尾围栏，把中间内容原样作为一个
   `BlockCode` 发出。
4. 每次 emit 都把当前的 `Page`、`Section` 和单调递增的 `Order` 记到
   block 上，让后续阶段还原文档原始顺序。

关键是：parser 完全不知道 embedding、retrieval、prompt 这些事，它
只守住"bytes → []Block"这一边界。

## What Changed

- 新增 `agents/s02-parser/main.go`，里面是 `ParseMarkdown`。
- 这是流水线里第一个真正做数据变换的阶段；s01 只立了类型。

## Try It

```bash
go run ./agents/s02-parser
```

预期输出（截断）：

```
parsed 6 blocks
  [heading] page=1 section="RAG-Anything (mini)" lang=    text="RAG-Anything (mini)"
  [text]    page=1 section="RAG-Anything (mini)" lang=    text="Retrieval-Augmented Generation lets an LLM answer u…"
  [heading] page=1 section="Pipeline"             lang=    text="Pipeline"
  [text]    page=1 section="Pipeline"             lang=    text="The pipeline has four stages: parse, embed, retri…"
  [code]    page=1 section="Pipeline"             lang=go  text="chunks := store.Search(rag.Embed(query), 5)"
  [heading] page=2 section="Multi-modal"          lang=    text="Multi-modal"
  ...
```

测试：

```bash
go test ./agents/s02-parser
```

## Upstream Source Reading

上游 `raganything/parser.py` 里有一个 `Parser` 基类，把
`OFFICE_FORMATS`、`IMAGE_FORMATS`、`TEXT_FORMATS` 这些格式常量摆好，
然后按文件类型派发到 MinerU / Docling / PaddleOCR。导读节选见
[`upstream-readings/s02-parser.go.md`](../../upstream-readings/s02-parser.go.md)，
里面解释了它的格式路由表，以及上游输出本质上就是一个 "list of
markdown blocks"，每个 block 带 `type`、`page_idx`、`captions`、
`footnotes` —— 这跟我们 `[]rag.Block` 的形状已经非常接近。
