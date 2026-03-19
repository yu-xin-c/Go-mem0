package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go-mem0/mem0"
)

func main() {
	// MEM0_ADDR 用于指定监听地址，例如 ":8080" 或 "127.0.0.1:8080"。
	addr := os.Getenv("MEM0_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	extractor := mem0.KeywordExtractor{}
	embedder, err := mem0.NewHashEmbedder(384)
	if err != nil {
		log.Fatal(err)
	}
	store := mem0.NewInMemoryStore()
	manager, err := mem0.NewManager(embedder, store)
	if err != nil {
		log.Fatal(err)
	}
	planner, err := buildPlanner(extractor)
	if err != nil {
		log.Fatal(err)
	}
	service, err := mem0.NewService(manager, planner)
	if err != nil {
		log.Fatal(err)
	}
	server, err := mem0.NewServer(service)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("mem0d listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func buildPlanner(extractor mem0.Extractor) (mem0.Planner, error) {
	baseURL := os.Getenv("MEM0_LLM_BASE_URL")
	apiKey := os.Getenv("MEM0_LLM_API_KEY")
	model := os.Getenv("MEM0_LLM_MODEL")
	if strings.TrimSpace(baseURL) != "" && strings.TrimSpace(apiKey) != "" && strings.TrimSpace(model) != "" {
		client, err := mem0.NewOpenAIClient(baseURL, apiKey)
		if err != nil {
			return nil, err
		}
		return mem0.NewLLMPlanner(client, model)
	}
	return mem0.NewRulePlanner(extractor)
}
