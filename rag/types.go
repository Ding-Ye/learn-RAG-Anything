// Package rag holds the shared, behavior-free type vocabulary used across
// every chapter of the learn-RAG-Anything curriculum.
//
// Each chapter defines its OWN concrete behavior in its own directory;
// this package only declares the data shapes that flow between stages,
// so that any chapter can read like a self-contained program while still
// composing cleanly with the others.
package rag

// BlockKind labels the modality of a Block. The upstream framework
// (HKUDS/RAG-Anything) supports text, image, table, equation, and code;
// in this teaching repo we focus on text, heading and code, plus stubs
// for the multimodal kinds so the appendix discussion has anchors.
type BlockKind string

const (
	BlockText     BlockKind = "text"
	BlockHeading  BlockKind = "heading"
	BlockCode     BlockKind = "code"
	BlockImage    BlockKind = "image"
	BlockTable    BlockKind = "table"
	BlockEquation BlockKind = "equation"
)

// Block is one typed unit emerging from a parser. Pages and ordering are
// retained because retrieval cares about source-location metadata.
type Block struct {
	Kind    BlockKind         `json:"kind"`
	Text    string            `json:"text"`
	Page    int               `json:"page,omitempty"`
	Order   int               `json:"order"`
	Section string            `json:"section,omitempty"`
	Lang    string            `json:"lang,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// Chunk is the unit a retriever returns. It is normally produced by
// concatenating several Blocks until a soft size budget is hit, with a
// configured overlap to keep semantic continuity across boundaries.
type Chunk struct {
	ID      string            `json:"id"`
	Text    string            `json:"text"`
	DocID   string            `json:"doc_id"`
	Section string            `json:"section,omitempty"`
	Pages   []int             `json:"pages,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// Embedding is a dense vector representation of either a chunk or a
// query. In real systems this is produced by a model; in this teaching
// repo it is a deterministic hash-derived float vector (see s04).
type Embedding []float32

// VectorRecord is the row a vector store keeps for each chunk.
type VectorRecord struct {
	ChunkID   string
	Embedding Embedding
	Chunk     Chunk
}

// RetrievedChunk is what a Retriever returns: the chunk plus the score
// it earned against the query. Scores are similarity (higher is better).
type RetrievedChunk struct {
	Chunk Chunk
	Score float32
}

// Document groups blocks under a single source identifier so the
// downstream stages can preserve provenance.
type Document struct {
	ID     string
	Source string
	Blocks []Block
}
