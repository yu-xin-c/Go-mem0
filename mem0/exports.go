package mem0

import (
	"time"

	"go-mem0/mem0/api"
	"go-mem0/mem0/embedder"
	"go-mem0/mem0/extractor"
	"go-mem0/mem0/llm"
	"go-mem0/mem0/manager"
	"go-mem0/mem0/model"
	"go-mem0/mem0/planner"
	"go-mem0/mem0/service"
	"go-mem0/mem0/store"
)

type Message = model.Message
type Memory = model.Memory
type MemoryCandidate = model.MemoryCandidate
type ScoredMemory = model.ScoredMemory
type MemoryOp = model.MemoryOp

type Embedder = embedder.Embedder
type HashEmbedder = embedder.HashEmbedder

func NewHashEmbedder(dim int) (*HashEmbedder, error) {
	return embedder.NewHashEmbedder(dim)
}

type Extractor = extractor.Extractor
type KeywordExtractor = extractor.KeywordExtractor

type Store = store.Store
type InMemoryStore = store.InMemoryStore

func NewInMemoryStore() *InMemoryStore {
	return store.NewInMemoryStore()
}

type Manager = manager.Manager
type ManagerOption = manager.Option

func WithNow(now func() time.Time) ManagerOption {
	return manager.WithNow(now)
}

func NewManager(embedder Embedder, store Store, opts ...ManagerOption) (*Manager, error) {
	return manager.New(embedder, store, opts...)
}

type Planner = planner.Planner
type RulePlanner = planner.RulePlanner
type LLMPlanner = planner.LLMPlanner

func NewRulePlanner(extractor Extractor) (*RulePlanner, error) {
	return planner.NewRulePlanner(extractor)
}

func NewLLMPlanner(client LLMClient, model string) (*LLMPlanner, error) {
	return planner.NewLLMPlanner(client, model)
}

type LLMMessage = llm.LLMMessage
type LLMClient = llm.LLMClient
type OpenAIClient = llm.OpenAIClient

func NewOpenAIClient(baseURL, apiKey string) (*OpenAIClient, error) {
	return llm.NewOpenAIClient(baseURL, apiKey)
}

type Service = service.Service

func NewService(manager *Manager, planner Planner) (*Service, error) {
	return service.New(manager, planner)
}

type Server = api.Server

func NewServer(service *Service) (*Server, error) {
	return api.NewServer(service)
}
