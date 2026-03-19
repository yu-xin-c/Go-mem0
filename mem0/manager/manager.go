package manager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"go-mem0/mem0/embedder"
	"go-mem0/mem0/model"
	"go-mem0/mem0/store"
)

type Manager struct {
	embedder embedder.Embedder
	store    store.Store
	now      func() time.Time
}

type Option func(*Manager)

func WithNow(now func() time.Time) Option {
	return func(m *Manager) {
		m.now = now
	}
}

func New(embedder embedder.Embedder, store store.Store, opts ...Option) (*Manager, error) {
	if embedder == nil {
		return nil, errors.New("embedder is nil")
	}
	if store == nil {
		return nil, errors.New("store is nil")
	}
	m := &Manager{
		embedder: embedder,
		store:    store,
		now:      time.Now,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}
	return m, nil
}

func (m *Manager) AddCandidates(ctx context.Context, userID string, candidates []model.MemoryCandidate) ([]model.Memory, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("userID is empty")
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	filtered := make([]model.MemoryCandidate, 0, len(candidates))
	input := make([]string, 0, len(candidates))
	for _, c := range candidates {
		text := strings.TrimSpace(c.Text)
		if text == "" {
			continue
		}
		filtered = append(filtered, c)
		input = append(input, text)
	}
	if len(input) == 0 {
		return nil, nil
	}
	embeddings, err := m.embedder.Embed(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(embeddings) != len(input) {
		return nil, errors.New("embedder returned mismatched embeddings length")
	}
	memories := make([]model.Memory, 0, len(input))
	for i, text := range input {
		mem, err := m.Upsert(ctx, userID, "", text, filtered[i].Metadata)
		if err != nil {
			return nil, err
		}
		memories = append(memories, mem)
	}
	return memories, nil
}

func (m *Manager) Upsert(ctx context.Context, userID, memoryID, text string, metadata map[string]string) (model.Memory, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return model.Memory{}, errors.New("userID is empty")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return model.Memory{}, errors.New("text is empty")
	}
	if memoryID == "" {
		memoryID = newMemoryID(userID, text, m.now())
	}
	mem := model.Memory{
		ID:        memoryID,
		UserID:    userID,
		Text:      text,
		CreatedAt: m.now().UTC(),
	}
	if metadata != nil {
		mem.Metadata = metadata
	}
	embs, err := m.embedder.Embed(ctx, []string{text})
	if err != nil {
		return model.Memory{}, err
	}
	if len(embs) != 1 {
		return model.Memory{}, errors.New("embedder returned invalid embedding count")
	}
	if err := m.store.Put(ctx, mem, embs[0]); err != nil {
		return model.Memory{}, err
	}
	return mem, nil
}

func (m *Manager) Search(ctx context.Context, userID, query string, limit int) ([]model.ScoredMemory, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("userID is empty")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is empty")
	}
	if limit <= 0 {
		limit = 10
	}
	embs, err := m.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(embs) != 1 {
		return nil, errors.New("embedder returned invalid embedding count")
	}
	return m.store.Search(ctx, userID, embs[0], limit)
}

func (m *Manager) List(ctx context.Context, userID string, limit int) ([]model.Memory, error) {
	return m.store.List(ctx, userID, limit)
}

func (m *Manager) Delete(ctx context.Context, userID, memoryID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("userID is empty")
	}
	memoryID = strings.TrimSpace(memoryID)
	if memoryID == "" {
		return errors.New("memoryID is empty")
	}
	return m.store.Delete(ctx, userID, memoryID)
}

func newMemoryID(userID, text string, now time.Time) string {
	sum := sha256.Sum256([]byte(userID + "\n" + text + "\n" + now.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(sum[:16])
}
