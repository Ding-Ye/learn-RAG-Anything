// Chapter s07 — prompt assembler.
//
// Once we have RetrievedChunks and a question, we must build the
// prompt the LLM will actually see. This is more design than code:
//
//   - the LLM should be told what is "context" and what is "question",
//   - context blocks should carry visible citations so the LLM can refer
//     to them in its answer,
//   - the assembled prompt should fit a budget; otherwise we silently
//     drop the lowest-scored chunk first.
//
// We model two prompt languages (EN and ZH) to mirror the upstream's
// bilingual prompt registry, and offer a single AssemblePrompt entry
// point.
//
// Run: go run ./agents/s07-prompt-assembler
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Ding-Ye/learn-RAG-Anything/rag"
)

// Lang is the prompt language. Adding more is one constant + one
// template each — exactly the pattern upstream uses with prompts.py /
// prompts_zh.py.
type Lang string

const (
	LangEN Lang = "en"
	LangZH Lang = "zh"
)

// AssembleOptions tunes a single prompt assembly call.
type AssembleOptions struct {
	Lang       Lang
	MaxChars   int     // truncate the joined context to fit; 0 = no cap
	MinScore   float32 // drop chunks below this similarity floor
	IncludeIDs bool    // print [c1] [c2] citation markers
}

// AssembledPrompt is the structured result. We return both the joined
// string (ready for an LLM call) and the chunks that survived
// truncation so callers can show citations to the user.
type AssembledPrompt struct {
	Text         string
	UsedChunks   []rag.RetrievedChunk
	DroppedCount int
}

// AssemblePrompt formats the prompt. It applies filters, sorts by
// score, drops low-scored chunks, then truncates from the *bottom*
// (lowest score first) until the joined context fits MaxChars.
func AssemblePrompt(question string, hits []rag.RetrievedChunk, opts AssembleOptions) AssembledPrompt {
	if opts.Lang == "" {
		opts.Lang = LangEN
	}

	keep := make([]rag.RetrievedChunk, 0, len(hits))
	for _, h := range hits {
		if h.Score >= opts.MinScore {
			keep = append(keep, h)
		}
	}
	sort.SliceStable(keep, func(i, j int) bool { return keep[i].Score > keep[j].Score })

	// Truncate from the bottom if MaxChars is set.
	dropped := 0
	if opts.MaxChars > 0 {
		for charsOf(keep, opts.IncludeIDs) > opts.MaxChars && len(keep) > 0 {
			keep = keep[:len(keep)-1]
			dropped++
		}
	}

	return AssembledPrompt{
		Text:         renderTemplate(opts.Lang, question, keep, opts.IncludeIDs),
		UsedChunks:   keep,
		DroppedCount: dropped,
	}
}

// charsOf returns the rough char-count of the formatted context only
// (not including the surrounding template), so MaxChars caps the part
// that actually grows with retrieval depth.
func charsOf(hits []rag.RetrievedChunk, ids bool) int {
	n := 0
	for i, h := range hits {
		if ids {
			n += len(fmt.Sprintf("[c%d] ", i+1))
		}
		n += len(h.Chunk.Text) + 1
	}
	return n
}

// renderTemplate produces the final prompt string. Two languages,
// otherwise identical structure. The structure on purpose mirrors what
// hosted-LLM RAG tutorials universally recommend: a clear role, a
// labeled context block, and a clearly delimited question.
func renderTemplate(lang Lang, question string, hits []rag.RetrievedChunk, ids bool) string {
	var b strings.Builder
	switch lang {
	case LangZH:
		b.WriteString("你是一个严谨的助手。请仅依据【参考资料】回答【问题】，")
		b.WriteString("如果资料里没有答案，明确说出『资料里没有提及』。\n\n")
		b.WriteString("【参考资料】\n")
	default:
		b.WriteString("You are a careful assistant. Answer the QUESTION using ONLY the CONTEXT below. ")
		b.WriteString("If the answer is not in the context, say \"the context does not say\".\n\n")
		b.WriteString("CONTEXT:\n")
	}
	for i, h := range hits {
		prefix := ""
		if ids {
			prefix = fmt.Sprintf("[c%d] ", i+1)
		}
		var meta string
		if h.Chunk.Section != "" {
			meta = fmt.Sprintf(" (section=%s)", h.Chunk.Section)
		}
		fmt.Fprintf(&b, "%s%s%s\n", prefix, strings.TrimSpace(h.Chunk.Text), meta)
	}
	if lang == LangZH {
		fmt.Fprintf(&b, "\n【问题】\n%s\n\n【答案】\n", question)
	} else {
		fmt.Fprintf(&b, "\nQUESTION:\n%s\n\nANSWER:\n", question)
	}
	return b.String()
}

func main() {
	hits := []rag.RetrievedChunk{
		{Chunk: rag.Chunk{ID: "c1", Text: "Retrieval-Augmented Generation grounds an LLM in external documents.", Section: "Intro"}, Score: 0.83},
		{Chunk: rag.Chunk{ID: "c2", Text: "Chunks are embedded as dense vectors and stored in a vector index.", Section: "Pipeline"}, Score: 0.71},
		{Chunk: rag.Chunk{ID: "c3", Text: "A retriever returns the top-k chunks for a query.", Section: "Pipeline"}, Score: 0.66},
		{Chunk: rag.Chunk{ID: "c4", Text: "Off-topic chunk about cats and mats.", Section: "Other"}, Score: 0.05},
	}

	en := AssemblePrompt("What does RAG actually do?", hits, AssembleOptions{
		Lang: LangEN, MinScore: 0.2, IncludeIDs: true, MaxChars: 220,
	})
	fmt.Println("--- EN prompt ---")
	fmt.Println(en.Text)
	fmt.Printf("(used=%d, dropped=%d)\n", len(en.UsedChunks), en.DroppedCount)

	zh := AssemblePrompt("RAG 实际上在做什么？", hits, AssembleOptions{
		Lang: LangZH, MinScore: 0.2, IncludeIDs: true,
	})
	fmt.Println("\n--- ZH prompt ---")
	fmt.Println(zh.Text)
}
