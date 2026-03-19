package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMClient interface {
	Chat(ctx context.Context, model string, messages []LLMMessage) (string, error)
}

type OpenAIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIClient(baseURL, apiKey string) (*OpenAIClient, error) {
	baseURL = strings.TrimSpace(baseURL)
	apiKey = strings.TrimSpace(apiKey)
	if baseURL == "" {
		return nil, errors.New("baseURL is empty")
	}
	if apiKey == "" {
		return nil, errors.New("apiKey is empty")
	}
	return &OpenAIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *OpenAIClient) Chat(ctx context.Context, model string, messages []LLMMessage) (string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", errors.New("model is empty")
	}
	reqBody := map[string]any{
		"model":       model,
		"messages":    messages,
		"temperature": 0,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := strings.TrimRight(c.baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("llm request failed")
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("llm response empty")
	}
	return parsed.Choices[0].Message.Content, nil
}
