# s02 · parser

Convert a markdown source string into typed `rag.Block`s. We handle
headings, fenced code, paragraphs, and a `---` page-break marker.

```bash
go run ./agents/s02-parser
go test  ./agents/s02-parser
```

Explainer: [`docs/en/s02-parser.md`](../../docs/en/s02-parser.md) ·
[`docs/zh/s02-parser.md`](../../docs/zh/s02-parser.md).
