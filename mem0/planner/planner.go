package planner

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"go-mem0/mem0/extractor"
	"go-mem0/mem0/llm"
	"go-mem0/mem0/model"
)

type Planner interface {
	Plan(ctx context.Context, userID string, messages []model.Message, existing []model.Memory) ([]model.MemoryOp, error)
}

type RulePlanner struct {
	extractor extractor.Extractor
}

func NewRulePlanner(extractor extractor.Extractor) (*RulePlanner, error) {
	if extractor == nil {
		return nil, errors.New("extractor is nil")
	}
	return &RulePlanner{extractor: extractor}, nil
}

func (r *RulePlanner) Plan(ctx context.Context, userID string, messages []model.Message, existing []model.Memory) ([]model.MemoryOp, error) {
	_ = existing
	candidates, err := r.extractor.Extract(ctx, userID, messages)
	if err != nil {
		return nil, err
	}
	ops := make([]model.MemoryOp, 0, len(candidates))
	for _, c := range candidates {
		text := strings.TrimSpace(c.Text)
		if text == "" {
			continue
		}
		ops = append(ops, model.MemoryOp{
			Kind:     "upsert",
			Text:     text,
			Metadata: c.Metadata,
		})
	}
	return ops, nil
}

type LLMPlanner struct {
	client llm.LLMClient
	model  string
}

func NewLLMPlanner(client llm.LLMClient, model string) (*LLMPlanner, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, errors.New("model is empty")
	}
	return &LLMPlanner{client: client, model: model}, nil
}

func (l *LLMPlanner) Plan(ctx context.Context, userID string, messages []model.Message, existing []model.Memory) ([]model.MemoryOp, error) {
	payload := map[string]any{
		"user_id":  userID,
		"messages": messages,
		"memories": existing,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	system := "You manage long-term memory for an AI assistant. Output JSON only."
	user := "Given the input, return JSON: {\"ops\":[{\"kind\":\"upsert\",\"memory_id\":\"optional\",\"text\":\"...\",\"metadata\":{}}]}. Use kind \"upsert\" or \"delete\". Input:\n" + string(raw)
	content, err := l.client.Chat(ctx, l.model, []llm.LLMMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	})
	if err != nil {
		return nil, err
	}
	jsonBody := extractJSON(content)
	if jsonBody == "" {
		return nil, errors.New("llm response missing json")
	}
	var resp struct {
		Ops []model.MemoryOp `json:"ops"`
	}
	if err := json.Unmarshal([]byte(jsonBody), &resp); err != nil {
		return nil, err
	}
	ops := make([]model.MemoryOp, 0, len(resp.Ops))
	for _, op := range resp.Ops {
		kind := strings.ToLower(strings.TrimSpace(op.Kind))
		if kind != "upsert" && kind != "delete" {
			continue
		}
		op.Kind = kind
		op.Text = strings.TrimSpace(op.Text)
		if op.Kind == "upsert" && op.Text == "" {
			continue
		}
		ops = append(ops, op)
	}
	return ops, nil
}

func extractJSON(input string) string {
	start := strings.Index(input, "{")
	end := strings.LastIndex(input, "}")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return input[start : end+1]
}
