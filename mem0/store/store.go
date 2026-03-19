package store

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"go-mem0/mem0/embedder"
	"go-mem0/mem0/model"
)

type Store interface {
	Put(ctx context.Context, mem model.Memory, embedding []float64) error
	Search(ctx context.Context, userID string, queryEmbedding []float64, limit int) ([]model.ScoredMemory, error)
	Delete(ctx context.Context, userID, memoryID string) error
	List(ctx context.Context, userID string, limit int) ([]model.Memory, error)
}

type InMemoryStore struct {
	mu    sync.RWMutex
	byUID map[string][]stored
}

type stored struct {
	mem model.Memory
	emb []float64
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		byUID: make(map[string][]stored),
	}
}

func (s *InMemoryStore) Put(ctx context.Context, mem model.Memory, embedding []float64) error {
	_ = ctx
	if mem.UserID == "" {
		return errors.New("memory has empty UserID")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.byUID[mem.UserID]
	for i := range items {
		if items[i].mem.ID == mem.ID {
			items[i].mem = mem
			items[i].emb = cloneVec(embedding)
			s.byUID[mem.UserID] = items
			return nil
		}
	}
	s.byUID[mem.UserID] = append(items, stored{mem: mem, emb: cloneVec(embedding)})
	return nil
}

func (s *InMemoryStore) Search(ctx context.Context, userID string, queryEmbedding []float64, limit int) ([]model.ScoredMemory, error) {
	_ = ctx
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userID is empty")
	}
	if limit <= 0 {
		limit = 10
	}
	s.mu.RLock()
	items := s.byUID[userID]
	s.mu.RUnlock()
	scored := make([]model.ScoredMemory, 0, len(items))
	for _, it := range items {
		scored = append(scored, model.ScoredMemory{
			Memory: it.mem,
			Score:  embedder.Cosine(queryEmbedding, it.emb),
		})
	}
	sort.Slice(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	if len(scored) > limit {
		scored = scored[:limit]
	}
	return scored, nil
}

func (s *InMemoryStore) Delete(ctx context.Context, userID, memoryID string) error {
	_ = ctx
	userID = strings.TrimSpace(userID)
	memoryID = strings.TrimSpace(memoryID)
	if userID == "" {
		return errors.New("userID is empty")
	}
	if memoryID == "" {
		return errors.New("memoryID is empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.byUID[userID]
	for i := range items {
		if items[i].mem.ID == memoryID {
			items[i] = items[len(items)-1]
			items = items[:len(items)-1]
			s.byUID[userID] = items
			return nil
		}
	}
	return nil
}

func (s *InMemoryStore) List(ctx context.Context, userID string, limit int) ([]model.Memory, error) {
	_ = ctx
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userID is empty")
	}
	if limit <= 0 {
		limit = 50
	}
	s.mu.RLock()
	items := s.byUID[userID]
	s.mu.RUnlock()
	out := make([]model.Memory, 0, len(items))
	for _, it := range items {
		out = append(out, it.mem)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func cloneVec(v []float64) []float64 {
	if v == nil {
		return nil
	}
	cp := make([]float64, len(v))
	copy(cp, v)
	return cp
}
