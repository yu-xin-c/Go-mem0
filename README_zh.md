# Go-mem0

一个为 AI 系统设计的记忆管理插件，灵感来源于 mem0 框架。该项目提供了一个轻量级、可嵌入的记忆层，可作为 OpenClaw 或其他 AI 系统的插件使用。

## 功能特性

- **记忆管理**：存储、检索和更新记忆，支持语义搜索
- **LLM 集成**：使用 LLM 根据对话上下文规划记忆操作（更新/删除）
- **规则回退**：当未配置 LLM 时，使用基于规则的规划器
- **向量嵌入**：内置基于哈希的向量嵌入，用于语义相似度计算
- **HTTP API**：提供 RESTful 接口用于记忆操作
- **线程安全存储**：内存存储，支持线程安全操作
- **可扩展架构**：模块化设计，清晰的职责分离

## 架构设计

项目采用模块化架构，包含以下组件：

- **model**：核心数据结构和类型定义
- **embedder**：文本到向量的转换
- **extractor**：从对话中提取记忆候选
- **store**：记忆持久化和检索
- **manager**：协调记忆操作
- **planner**：规划记忆操作（基于规则或 LLM）
- **llm**：用于记忆规划的 LLM 客户端
- **service**：插件能力的服务层
- **api**：HTTP 服务器和路由

## 安装指南

### 前提条件
- Go 1.22+
-（可选）LLM API 访问权限，用于基于 LLM 的记忆规划

### 安装步骤

1. 克隆仓库：

```bash
git clone https://github.com/yourusername/go-mem0.git
cd go-mem0
```

2. 构建项目：

```bash
go build -o mem0d ./cmd/mem0d
```

3. 运行服务器：

```bash
./mem0d
```

## 配置选项

服务器可通过环境变量进行配置：

| 环境变量 | 描述 | 默认值 |
|---------|------|--------|
| `MEM0_ADDR` | 服务器地址和端口 | `:8080` |
| `MEM0_LLM_BASE_URL` | LLM API 基础 URL（例如 `https://api.openai.com`） | - |
| `MEM0_LLM_API_KEY` | LLM API 密钥 | - |
| `MEM0_LLM_MODEL` | LLM 模型名称（例如 `gpt-3.5-turbo`） | - |

## API 接口

### 更新记忆

**POST /v1/memories/update**

根据对话上下文更新记忆。规划器（基于规则或 LLM）将生成适当的记忆操作。

**请求体：**

```json
{
  "user_id": "user123",
  "messages": [
    {
      "role": "user",
      "content": "我喜欢喝拿铁。我的工作城市是上海。"
    }
  ]
}
```

**响应：**

```json
{
  "memories": [
    {
      "id": "abc123",
      "user_id": "user123",
      "text": "我喜欢喝拿铁。",
      "created_at": "2026-03-19T10:00:00Z"
    },
    {
      "id": "def456",
      "user_id": "user123",
      "text": "我的工作城市是上海。",
      "created_at": "2026-03-19T10:00:00Z"
    }
  ]
}
```

### 查询记忆

**GET /v1/memories/search**

通过语义相似度搜索记忆。

**查询参数：**
- `user_id`：用户标识符（必需）
- `q`：搜索查询（必需）
- `limit`：最大结果数（默认：10）

**响应：**

```json
{
  "results": [
    {
      "memory": {
        "id": "abc123",
        "user_id": "user123",
        "text": "我喜欢喝拿铁。",
        "created_at": "2026-03-19T10:00:00Z"
      },
      "score": 0.95
    }
  ]
}
```

### 查看记忆列表

**GET /v1/memories**

列出用户的记忆，按创建时间排序（最新优先）。

**查询参数：**
- `user_id`：用户标识符（必需）
- `limit`：最大结果数（默认：50）

**响应：**

```json
{
  "memories": [
    {
      "id": "def456",
      "user_id": "user123",
      "text": "我的工作城市是上海。",
      "created_at": "2026-03-19T10:00:00Z"
    },
    {
      "id": "abc123",
      "user_id": "user123",
      "text": "我喜欢喝拿铁。",
      "created_at": "2026-03-19T10:00:00Z"
    }
  ]
}
```

## 使用示例

### 基本使用

```go
package main

import (
	"context"
	"fmt"
	"go-mem0/mem0"
)

func main() {
	// 创建组件
	extractor := mem0.KeywordExtractor{}
	embedder, _ := mem0.NewHashEmbedder(384)
	store := mem0.NewInMemoryStore()
	manager, _ := mem0.NewManager(embedder, store)
	planner, _ := mem0.NewRulePlanner(extractor)
	service, _ := mem0.NewService(manager, planner)

	// 从消息更新记忆
	messages := []mem0.Message{
		{Role: "user", Content: "我喜欢喝拿铁。我的工作城市是上海。"},
	}
	memories, _ := service.UpdateFromMessages(context.Background(), "user123", messages)
	fmt.Printf("创建了 %d 条记忆\n", len(memories))

	// 搜索记忆
	results, _ := service.Search(context.Background(), "user123", "咖啡", 5)
	fmt.Printf("找到 %d 个结果\n", len(results))
	for _, result := range results {
		fmt.Printf("得分: %.2f, 文本: %s\n", result.Score, result.Memory.Text)
	}

	// 列出记忆
	allMemories, _ := service.List(context.Background(), "user123", 50)
	fmt.Printf("总记忆数: %d\n", len(allMemories))
}
```

### 使用 LLM 规划器

```go
package main

import (
	"context"
	"fmt"
	"go-mem0/mem0"
)

func main() {
	// 创建组件
	extractor := mem0.KeywordExtractor{}
	embedder, _ := mem0.NewHashEmbedder(384)
	store := mem0.NewInMemoryStore()
	manager, _ := mem0.NewManager(embedder, store)

	// 创建 LLM 客户端和规划器
	llmClient, _ := mem0.NewOpenAIClient("https://api.openai.com", "your-api-key")
	planner, _ := mem0.NewLLMPlanner(llmClient, "gpt-3.5-turbo")

	service, _ := mem0.NewService(manager, planner)

	// 使用 LLM 规划从消息更新记忆
	messages := []mem0.Message{
		{Role: "user", Content: "我喜欢喝拿铁。我的工作城市是上海。"},
	}
	memories, _ := service.UpdateFromMessages(context.Background(), "user123", messages)
	fmt.Printf("创建了 %d 条记忆\n", len(memories))
}
```

## 作为服务运行

### 启动服务器

```bash
# 使用默认设置
./mem0d

# 使用自定义地址
MEM0_ADDR=:9090 ./mem0d

# 启用 LLM 集成
MEM0_LLM_BASE_URL=https://api.openai.com MEM0_LLM_API_KEY=your-api-key MEM0_LLM_MODEL=gpt-3.5-turbo ./mem0d
```

### 健康检查

```bash
curl http://localhost:8080/healthz
# 响应: {"status":"ok"}
```

## 测试

运行测试：

```bash
GOTOOLCHAIN=go1.25.7 go test ./...
```

## 贡献指南

欢迎贡献！请随时提交 Pull Request。

### 开发规范

1. 遵循 Go 代码规范
2. 为新功能编写测试
3. 根据需要更新文档
4. 保持提交消息清晰简洁

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- 受 [mem0](https://github.com/mem0ai/mem0) 框架启发
- 使用 Go 标准库和最佳实践构建
