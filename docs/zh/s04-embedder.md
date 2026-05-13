# s04 · embedder

## Problem

检索需要一种可比较的"语义"表达。业界标准答案是 *embedding*：把
文本映射到稠密向量，让语义相近的文本在向量空间里靠近。真实系统通常
是调用一个云端模型（OpenAI / Cohere / Voyage 等）——要花钱、要网络、
还把测试绑死在某个 API key 上。教学仓库需要：

- 完全不联网（CI 离线也能过），
- 同一段输入，在任何机器上都得到同一组向量，
- 同时还要让相似度*真的有意义*，否则下一章的检索 demo 全是噪声。

## Solution

把 `Embedder` 定成一个非常小的 Go 接口：

```go
type Embedder interface {
    Dim() int
    Embed(text string) rag.Embedding
}
```

然后实现 `FakeEmbedder`：按空白和常见标点分词，对每个 token 做
SHA-256，把哈希字节拆成 4 字节窗口，分别投到固定维度向量的某个位置
上，并按符号位决定 `+/-` 贡献；最后做 L2 归一化。共享 token 多的两段
文本，在余弦空间里就会落得近 —— 不是因为模型"懂"，而是因为它们
散射到了重叠的位置上。

后面所有章节依赖的是这个接口。日后接真实模型，只改一个文件即可。

## How It Works

对每个 token `t`：

1. `sum = SHA256(t)` —— 32 字节的确定性噪声。
2. 把 `sum` 按 4 字节窗口扫一遍。每个窗口决定一个位置
   `p ∈ [0, dim)`，并按最高位决定一个 +/- 符号。
3. 给 `vec[p]` 累加 `±1/len(tokens)`。除以 token 数让不同长度输入的
   幅值大致可比。
4. 全部累加完后，`l2Normalize(vec)`：归一化之后，两个向量的余弦相似度
   就等于它们的点积，s05 会用到。

`CosineSimilarity` 在这里就先实现了，s05 的 top-k 检索会直接用。

## What Changed

- `rag.Embedding` 第一次有了非平凡的产生者。
- 引入接口边界：所有后续章节都接收 `Embedder` 接口而非具体类型，
  之后接真实模型可以一行换掉，不影响 pipeline 别的部分。

## Try It

```bash
go run ./agents/s04-embedder
```

预期：

```
Embedder dim=64
  sim=+0.430  "retrieval augmented generation"  vs  "RAG combines retrieval with generation"
  sim=-0.120  "retrieval augmented generation"  vs  "the cat sat on the mat"
  sim=+1.000  "hello world"  vs  "hello world"
```

匹配的两段 >0；无关的两段 ≈0 或略负；完全相同的两段 = 1.0。
这正是检索想要的性质。

测试：

```bash
go test ./agents/s04-embedder
```

## Upstream Source Reading

上游把 embedder 当作*依赖注入的 callable*：`raganything/raganything.py`
接收一个 `embedding_func` 参数，传给 LightRAG，由 LightRAG 在分块/
建索引的循环里调用。具体模型由用户自己接（OpenAI 等）。导读见
[`upstream-readings/s04-embedder.go.md`](../../upstream-readings/s04-embedder.go.md)，
里面有上游片段与我们 `Embedder` 接口的对照。
