# Go-mem0

A memory management plugin for AI systems, inspired by the mem0 framework. This project provides a lightweight, embeddable memory layer that can be used as a plugin for OpenClaw or other AI systems.

## Features

- **Memory Management**: Store, retrieve, and update memories with semantic search capabilities
- **LLM Integration**: Use LLM to plan memory operations (upsert/delete) based on conversation context
- **Rule-based Fallback**: Default rule-based planner when LLM is not configured
- **Vector Embedding**: Built-in hash-based vector embedding for semantic similarity
- **HTTP API**: RESTful endpoints for memory operations
- **Thread-safe Storage**: In-memory storage with thread-safe operations
- **Extensible Architecture**: Modular design with clear separation of concerns

## Architecture

The project follows a modular architecture with the following components:

- **model**: Core data structures and types
- **embedder**: Text-to-vector conversion
- **extractor**: Extract memory candidates from conversations
- **store**: Memory persistence and retrieval
- **manager**: Orchestrates memory operations
- **planner**: Plans memory operations (rule-based or LLM-based)
- **llm**: LLM client for memory planning
- **service**: Service layer for plugin capabilities
- **api**: HTTP server and routes

## Installation

### Prerequisites
- Go 1.22+
- (Optional) LLM API access for LLM-based memory planning

### Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/go-mem0.git
cd go-mem0
```

2. Build the project:

```bash
go build -o mem0d ./cmd/mem0d
```

3. Run the server:

```bash
./mem0d
```

## Configuration

The server can be configured using environment variables:

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `MEM0_ADDR` | Server address and port | `:8080` |
| `MEM0_LLM_BASE_URL` | LLM API base URL (e.g., `https://api.openai.com`) | - |
| `MEM0_LLM_API_KEY` | LLM API key | - |
| `MEM0_LLM_MODEL` | LLM model name (e.g., `gpt-3.5-turbo`) | - |

## API Endpoints

### Update Memory

**POST /v1/memories/update**

Update memory based on conversation context. The planner (rule-based or LLM-based) will generate appropriate memory operations.

**Request Body:**

```json
{
  "user_id": "user123",
  "messages": [
    {
      "role": "user",
      "content": "I like drinking latte. My working city is Shanghai."
    }
  ]
}
```

**Response:**

```json
{
  "memories": [
    {
      "id": "abc123",
      "user_id": "user123",
      "text": "I like drinking latte.",
      "created_at": "2026-03-19T10:00:00Z"
    },
    {
      "id": "def456",
      "user_id": "user123",
      "text": "My working city is Shanghai.",
      "created_at": "2026-03-19T10:00:00Z"
    }
  ]
}
```

### Search Memory

**GET /v1/memories/search**

Search for memories by semantic similarity.

**Query Parameters:**
- `user_id`: User identifier (required)
- `q`: Search query (required)
- `limit`: Maximum number of results (default: 10)

**Response:**

```json
{
  "results": [
    {
      "memory": {
        "id": "abc123",
        "user_id": "user123",
        "text": "I like drinking latte.",
        "created_at": "2026-03-19T10:00:00Z"
      },
      "score": 0.95
    }
  ]
}
```

### List Memories

**GET /v1/memories**

List memories for a user, sorted by creation time (newest first).

**Query Parameters:**
- `user_id`: User identifier (required)
- `limit`: Maximum number of results (default: 50)

**Response:**

```json
{
  "memories": [
    {
      "id": "def456",
      "user_id": "user123",
      "text": "My working city is Shanghai.",
      "created_at": "2026-03-19T10:00:00Z"
    },
    {
      "id": "abc123",
      "user_id": "user123",
      "text": "I like drinking latte.",
      "created_at": "2026-03-19T10:00:00Z"
    }
  ]
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"go-mem0/mem0"
)

func main() {
	// Create components
	extractor := mem0.KeywordExtractor{}
	embedder, _ := mem0.NewHashEmbedder(384)
	store := mem0.NewInMemoryStore()
	manager, _ := mem0.NewManager(embedder, store)
	planner, _ := mem0.NewRulePlanner(extractor)
	service, _ := mem0.NewService(manager, planner)

	// Update memory from messages
	messages := []mem0.Message{
		{Role: "user", Content: "I like drinking latte. My working city is Shanghai."},
	}
	memories, _ := service.UpdateFromMessages(context.Background(), "user123", messages)
	fmt.Printf("Created %d memories\n", len(memories))

	// Search memory
	results, _ := service.Search(context.Background(), "user123", "coffee", 5)
	fmt.Printf("Found %d results\n", len(results))
	for _, result := range results {
		fmt.Printf("Score: %.2f, Text: %s\n", result.Score, result.Memory.Text)
	}

	// List memories
	allMemories, _ := service.List(context.Background(), "user123", 50)
	fmt.Printf("Total memories: %d\n", len(allMemories))
}
```

### Using LLM Planner

```go
package main

import (
	"context"
	"fmt"
	"go-mem0/mem0"
)

func main() {
	// Create components
	extractor := mem0.KeywordExtractor{}
	embedder, _ := mem0.NewHashEmbedder(384)
	store := mem0.NewInMemoryStore()
	manager, _ := mem0.NewManager(embedder, store)

	// Create LLM client and planner
	llmClient, _ := mem0.NewOpenAIClient("https://api.openai.com", "your-api-key")
	planner, _ := mem0.NewLLMPlanner(llmClient, "gpt-3.5-turbo")

	service, _ := mem0.NewService(manager, planner)

	// Update memory from messages using LLM planning
	messages := []mem0.Message{
		{Role: "user", Content: "I like drinking latte. My working city is Shanghai."},
	}
	memories, _ := service.UpdateFromMessages(context.Background(), "user123", messages)
	fmt.Printf("Created %d memories\n", len(memories))
}
```

## Running as a Service

### Start the server

```bash
# With default settings
./mem0d

# With custom address
MEM0_ADDR=:9090 ./mem0d

# With LLM integration
MEM0_LLM_BASE_URL=https://api.openai.com MEM0_LLM_API_KEY=your-api-key MEM0_LLM_MODEL=gpt-3.5-turbo ./mem0d
```

### Health Check

```bash
curl http://localhost:8080/healthz
# Response: {"status":"ok"}
```

## Testing

Run the tests:

```bash
GOTOOLCHAIN=go1.25.7 go test ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Guidelines

1. Follow Go code conventions
2. Write tests for new functionality
3. Update documentation as needed
4. Keep commit messages clear and concise

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- Inspired by the [mem0](https://github.com/mem0ai/mem0) framework
- Built with Go's standard library and best practices
