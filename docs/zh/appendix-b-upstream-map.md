# 附录 B · 上游文件地图

## Problem

当你从本教学仓库跳到真实上游
`HKUDS/RAG-Anything @ 146828f7…`，会看到 19 个 Python 文件、约
12K 行代码。没有地图就不知道先打哪里。本附录就是这张地图。

## Solution

一张表：把所有我们关心的上游文件、它各自负责什么、对应到本仓库
哪一章列清楚。把它放在副屏，阅读上游时随时对照。

## How It Works

上游 `raganything/` 下的全景：

| 文件                     | LOC   | 负责                                                                     | 对应章节       |
| ------------------------ | ----: | ------------------------------------------------------------------------ | -------------- |
| `__init__.py`            |   114 | 对外接口；re-export `RAGAnything`                                         | —              |
| `base.py`                |    12 | `DocStatus` 枚举                                                          | s01 词表       |
| `config.py`              |   158 | `RAGAnythingConfig` dataclass + 环境变量加载                              | s08 表面       |
| `parser.py`              |  2660 | `Parser` 基类 + MinerU / Docling / PaddleOCR 适配 + CLI                   | s02            |
| `enhanced_markdown.py`   |   534 | 解析后的 markdown 后处理（链接、资源 URL 等）                              | s02（深入）    |
| `asset_urls.py`          |   117 | 给图像 / 表格附公开 URL                                                   | s02（深入）    |
| `processor.py`           |  2229 | `ProcessorMixin`：分块、入库、实体/关系抽取                                | s03、s08       |
| `modalprocessors.py`     |  1607 | 图像 / 表格 / 公式 / 通用模态处理器 + `ContextExtractor`                    | 附录 A         |
| `omml_extractor.py`      |   758 | OMML (Office 数学公式) → LaTeX 转换                                       | 附录 A         |
| `utils.py`               |   380 | helper：`get_processor_for_type`、`get_table_body`、图片 base64 等         | s01、s03、附录 A |
| `query.py`               |   868 | `QueryMixin`：文本/多模态查询、缓存键                                       | s06            |
| `prompt.py`              |   406 | `PromptRegistry` + 英文 prompt 模板                                        | s07            |
| `prompts_zh.py`          |   337 | 中文 prompt 翻译                                                          | s07            |
| `prompt_manager.py`      |   156 | 语言切换、按调用覆盖                                                       | s07            |
| `resilience.py`          |   397 | 重试 / 超时原语（有界重试、指数退避）                                       | （未来 s09）   |
| `batch.py`               |   428 | `BatchMixin`：处理 N 份文档并报告进度                                       | （未来 s10）   |
| `batch_parser.py`        |   470 | 独立批量解析 CLI                                                          | （未来 s10）   |
| `callbacks.py`           |   377 | 长流程的进度 / 阶段回调                                                    | （未来）       |
| `raganything.py`         |   644 | 编排器类，组合所有 mixin                                                   | s08            |

合计：约 12.6K 行 Python。

## What Changed

代码上没有任何变化。本附录是参考索引，建议和
`upstream-readings/sNN-*.go.md` 一起读，那里有更深入的章节级节选。

## Try It

如果想快速过一遍上游，推荐这个顺序：

1. `__init__.py` —— 看对外接口。
2. `config.py` —— 看 RAGAnything 暴露了哪些旋钮。
3. `raganything.py` —— 看编排类和构造函数。
4. `parser.py` —— 前半部分（类声明、格式常量）。
5. `processor.py` —— `_process_text_chunks` 这条线。
6. `modalprocessors.py` —— `ImageModalProcessor.process_modal` 入口。
7. `query.py` —— `aquery` 那条线。
8. `prompt.py` —— registry 机制 + 几个模板。

本仓库每章都对应其中一站。

## Upstream Source Reading

两个文件级的小贴士：

- **`parser.py` 很大**（2.6K 行），但大部分是 MinerU / Docling /
  PaddleOCR 的平台相关 shell-out 逻辑。真正有意思的 Python 在前
  约 200 行和 `MineruParser.parse` 里。
- **`processor.py` 也很大**（2.2K 行），主体是多模态协调。教学相关
  的代码主要是 `_process_text_chunks` 和 `_generate_cache_key`。

实用做法：开一个 `git grep` 窗口对着上游树，跟着每份
`upstream-readings/sNN-*.go.md` 按符号名跳转。
