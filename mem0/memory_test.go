package mem0

import (
	"context"
	"testing"
	"time"
)

func TestManagerAddAndSearch(t *testing.T) {
	extractor := KeywordExtractor{}
	embedder, err := NewHashEmbedder(64)
	if err != nil {
		t.Fatalf("embedder: %v", err)
	}
	store := NewInMemoryStore()
	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)
	manager, err := NewManager(embedder, store, WithNow(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	planner, err := NewRulePlanner(extractor)
	if err != nil {
		t.Fatalf("planner: %v", err)
	}
	service, err := NewService(manager, planner)
	if err != nil {
		t.Fatalf("service: %v", err)
	}
	msgs := []Message{{Role: "user", Content: "我喜欢喝拿铁。我的工作城市是上海。"}}
	memories, err := service.UpdateFromMessages(context.Background(), "u1", msgs)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if len(memories) != 2 {
		t.Fatalf("expected 2 memories, got %d", len(memories))
	}
	results, err := manager.Search(context.Background(), "u1", "喜欢 拿铁", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected results")
	}
}

func TestInMemoryStoreDelete(t *testing.T) {
	store := NewInMemoryStore()
	mem := Memory{ID: "m1", UserID: "u1", Text: "hello", CreatedAt: time.Now()}
	if err := store.Put(context.Background(), mem, []float64{1, 0, 0}); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := store.Delete(context.Background(), "u1", "m1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	results, err := store.Search(context.Background(), "u1", []float64{1, 0, 0}, 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %d", len(results))
	}
}
