package model

import "time"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Memory struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Text      string            `json:"text"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type MemoryCandidate struct {
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type ScoredMemory struct {
	Memory Memory  `json:"memory"`
	Score  float64 `json:"score"`
}

type MemoryOp struct {
	Kind     string            `json:"kind"`
	MemoryID string            `json:"memory_id,omitempty"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
