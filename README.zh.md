# learn-RAG-Anything

一个从零开始、动手实现的 Go 学习仓库，对应上游
[HKUDS/RAG-Anything](https://github.com/HKUDS/RAG-Anything) ——
分成 8 个小章节，每章用一份独立的 Go 程序，构建一条小而完整的
检索增强生成 (RAG) 流水线。

每章约 100 行代码，配双语讲解 + 上游 Python 源码导读。任何一章都
可以直接 `go run ./agents/sNN-*` 单独运行。

**English**: [README.md](./README.md).

## 课程目录

| #   | 章节                | 你将构建                                              | 运行                                |
| --- | ------------------- | ----------------------------------------------------- | ---------------------------------- |
| s01 | block-model         | 类型化 Block / Chunk 数据词表                          | `go run ./agents/s01-block-model`  |
| s02 | parser              | 极简 Markdown → []Block 解析器                         | `go run ./agents/s02-parser`       |
| s03 | chunker             | Blocks → 带元数据的重叠 Chunks                         | `go run ./agents/s03-chunker`      |
| s04 | embedder            | 基于哈希的确定性 Embedder（无需模型）                  | `go run ./agents/s04-embedder`     |
| s05 | vector-store        | 内存向量库 + 余弦相似度 top-k                          | `go run ./agents/s05-vector-store` |
| s06 | retriever           | Query → embed → top-k chunks                          | `go run ./agents/s06-retriever`    |
| s07 | prompt-assembler    | Chunks + 问题 → LLM Prompt                            | `go run ./agents/s07-prompt-assembler` |
| s08 | pipeline            | 端到端 parse → chunk → embed → store → answer        | `go run ./agents/s08-pipeline`     |

此外在 [`docs/zh`](./docs/zh) 中提供双语对照的章节讲解、一个端到端
追踪文档 ([`s_full-integration`](./docs/zh/s_full-integration.md))、
一篇关于多模态 RAG 的附录，以及一份上游文件索引。

## 仓库结构

```
rag/                     共享类型词表（Block、Chunk 等），无行为
agents/sNN-<name>/       每章一个 main.go + 测试 + README
docs/en/、docs/zh/       双语章节讲解，统一六段式
upstream-readings/       带注释的上游源码节选 (.go.md)
web/index.html           简易静态课程浏览器
.github/workflows/       CI：vet / build / test / 双语标题对齐检查
```

## 快速开始

```bash
git clone https://github.com/Ding-Ye/learn-RAG-Anything.git
cd learn-RAG-Anything
go test ./...
go run ./agents/s08-pipeline
```

## 六段式文档约定

每章讲解都遵循同样的六段结构，方便你按需要跳读：

1. **Problem** — 这一节要解决的真实问题。
2. **Solution** — 本章用大约 100 行 Go 给出的答案。
3. **How It Works** — 关键组件与数据流。
4. **What Changed** — 相比上一章新增了什么。
5. **Try It** — 具体命令与预期输出。
6. **Upstream Source Reading** — 对应上游 Python 文件的导读。

## 许可证与致谢

本仓库以 **Apache License 2.0** 发布（见 [LICENSE](./LICENSE)）。

它是出于教学目的的重新实现，参考自
[HKUDS/RAG-Anything](https://github.com/HKUDS/RAG-Anything) 在
`146828f73de652c9d72399bfc60499966f3f8bd0` 这一 commit 下的代码，上游为
MIT 协议。在此感谢上游作者。仓库内不直接复制上游源码；
`upstream-readings/` 目录下仅按学术引用习惯节选少量带注释的片段。

## 范围说明

这是一个**教学**仓库，不是生产级 RAG 框架。Embedder 用哈希、向量库
是内存切片、LLM 是占位实现。每处偷懒都在章节文档里点出，方便你日后
阅读真正的上游时分辨：哪些抽象是工程现实，哪些只是教学替身。
