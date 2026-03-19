package extractor

import (
	"context"
	"strings"

	"go-mem0/mem0/model"
)

type Extractor interface {
	Extract(ctx context.Context, userID string, messages []model.Message) ([]model.MemoryCandidate, error)
}

type KeywordExtractor struct{}

func (k KeywordExtractor) Extract(ctx context.Context, userID string, messages []model.Message) ([]model.MemoryCandidate, error) {
	_ = ctx
	_ = userID
	var out []model.MemoryCandidate
	for _, msg := range messages {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		if role != "user" {
			continue
		}
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			continue
		}
		lines := splitCandidates(text)
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if ln == "" {
				continue
			}
			out = append(out, model.MemoryCandidate{Text: ln})
		}
	}
	return out, nil
}

func splitCandidates(text string) []string {
	text = strings.ReplaceAll(text, "。", "\n")
	text = strings.ReplaceAll(text, "；", "\n")
	text = strings.ReplaceAll(text, ";", "\n")
	text = strings.ReplaceAll(text, ".", "\n")
	text = strings.ReplaceAll(text, "!", "\n")
	text = strings.ReplaceAll(text, "？", "\n")
	text = strings.ReplaceAll(text, "?", "\n")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(text, "\n")
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if len([]rune(p)) < 6 {
			continue
		}
		out = append(out, p)
	}
	return out
}
