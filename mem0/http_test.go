package mem0

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAddAndSearch(t *testing.T) {
	extractor := KeywordExtractor{}
	embedder, err := NewHashEmbedder(64)
	if err != nil {
		t.Fatalf("embedder: %v", err)
	}
	store := NewInMemoryStore()
	manager, err := NewManager(embedder, store)
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
	server, err := NewServer(service)
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	payload := struct {
		UserID   string    `json:"user_id"`
		Messages []Message `json:"messages"`
	}{
		UserID: "u2",
		Messages: []Message{
			{Role: "user", Content: "我在北京工作，喜欢跑步。"},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/memories/update", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("add status %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/memories/search?user_id=u2&q=跑步&limit=3", nil)
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search status %d", rec.Code)
	}
	var resp struct {
		Results []ScoredMemory `json:"results"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatalf("expected search results")
	}
}
