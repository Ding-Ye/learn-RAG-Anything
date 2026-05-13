# 附录 A · 多模态 RAG

## Problem

本课程构建的是*纯文本* pipeline。真实文档要复杂得多：一页里可能
混着叙述文字、表格、数学公式、配 caption 和脚注的插图。上游
`HKUDS/RAG-Anything` 之所以叫这个名字，正是因为它想成为"多模态文档
的全能 RAG"。所以问题来了：加入多模态到底会改变什么？我们为什么
没有在代码里实现它？

## Solution

附录把四种常见的非文本模态——**图像、表格、公式、OCR 字符**——
逐一拆开，每种解释：

1. 上游怎么表示它，
2. 多依赖了哪些外部系统，
3. 流水线在哪里分叉，
4. 哪些部分跟文本一致。

不增加 Go 代码。目标是让现有课程落地到现实，并明确告诉你：如果想
把这个仓库扩展成多模态版本，要改哪几章。

## How It Works

### 图像 (Image)

上游 `ImageModalProcessor`（`raganything/modalprocessors.py`）是个
小状态机：

1. **Caption（图像描述）。** 把图像传给**视觉 LLM**
   (`vision_model_func`)，配合 `raganything/prompt.py` 里的
   `vision_prompt_with_context` 模板。模型返回 JSON，里面有
   `detailed_description` 和 `entity_info`。
2. **embed 的是 caption，不是图像本身。** caption 走的是普通文本
   chunk 用的同一个 `embedding_func`，让检索统一。
3. **入库。** caption 变成一条 chunk-like 记录；图片路径作为
   metadata 保留，便于答案引用 chunk 时把图也展示出来。

如果在我们的 Go 代码里实现：加 `BlockKind = BlockImage`、加一个
`Captioner` 接口、在 chunker 里加一小段分发。后面所有阶段不变。

### 表格 (Table)

表格被*线性化*，不是当作图像 embed。上游 `TableModalProcessor` 分
两步：先用 `raganything/utils.py` 的 `get_table_body` 提取出清洗过
的 HTML / markdown，再可选地让 LLM 用 `TABLE_ANALYSIS_SYSTEM`
prompt 写一段自然语言摘要。摘要和原表都保留；embedding 用摘要
（因为用户提问通常是散文，不是表格单元格）。

### 公式 (Equation)

公式以 **TeX** 保存，从不渲染成图像。`EquationModalProcessor` 让
LLM 写一段对公式的"文字解释"（涉及哪些变量、常用在哪里、给一个
数值例子）。embedding 用这段解释，不是原始 TeX。

这里有个有意思的细节：Office 公式存的是 OMML 而不是 TeX，
`raganything/omml_extractor.py` 负责 OMML → TeX 的转换，跑模态处理
器之前先转。这种琐碎细节在教学仓库里会被吃掉，但生产里少不了。

### OCR 字符（扫描 PDF）

OCR 引擎（上游默认 PaddleOCR）抽出来的纯文本 block 带置信度。
pipeline 当普通文本处理，但用置信度对排序加权：置信度低的 OCR 文本
在进入向量库前就会被打折。我们没在 Go 里实现，如果要做，
合理位置是 `Block.Meta["confidence"]` 加 s05 Search 的钩子。

## What Changed

现有章节什么都没变。本附录只是*未来扩展*的地图。如果你想把这个
仓库扩展到多模态，最小改动配方是：

| 模态     | 新增内容                | 涉及章节                  |
| -------- | ----------------------- | ------------------------- |
| 图像     | `Captioner` 接口        | s02、s03、新增 `s09-image` |
| 表格     | `TableLinearizer`       | s02、s03                  |
| 公式     | TeX / OMML 解码         | s02、s03                  |
| OCR      | `Confidence` 元数据     | s02、s05                  |

所有模态最终都坍塌为"embedder 能看的一段文本"，所以向量库、
retriever、prompt assembler **不需要**改。这点洞察——多模态 RAG
本质是"生成文本的多种方式"，而不是"多种向量空间"——是本附录唯一的
要点。

## Try It

本附录不附带可运行代码，但你可以直接读上游：

```bash
less /path/to/upstream/raganything/modalprocessors.py
# 搜索: class ImageModalProcessor、class TableModalProcessor 等
less /path/to/upstream/raganything/omml_extractor.py
```

## Upstream Source Reading

| 模态     | 文件                              | 类 / 函数                    |
| -------- | --------------------------------- | ---------------------------- |
| 图像     | `raganything/modalprocessors.py`  | `ImageModalProcessor`        |
| 表格     | `raganything/modalprocessors.py`  | `TableModalProcessor`        |
| 公式     | `raganything/modalprocessors.py`  | `EquationModalProcessor`     |
| OMML→TeX | `raganything/omml_extractor.py`   | `extract_omml_to_latex`      |
| 通用     | `raganything/modalprocessors.py`  | `GenericModalProcessor`      |
| Prompts  | `raganything/prompt.py`           | `vision_prompt_with_context`、`TABLE_ANALYSIS_SYSTEM` 等 |
| Utility  | `raganything/utils.py`            | `get_table_body`、`get_equation_text_and_format` |
