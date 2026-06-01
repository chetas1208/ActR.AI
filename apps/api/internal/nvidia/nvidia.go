// Package nvidia provides a client for NVIDIA NIM OpenAI-compatible endpoints.
// NIM_API_KEY is required. NIM_BASE_URL defaults to https://integrate.api.nvidia.com/v1.
// All GenAI calls go through this package — no OpenAI SDK or OPENAI_* vars are used.
package nvidia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/config"
)

// Client calls NVIDIA NIM via the OpenAI-compatible /chat/completions endpoint.
type Client struct {
	apiKey    string
	baseURL   string
	llmModel  string
	fastModel string
	codeModel string
	http      *http.Client
}

func NewClient(cfg *config.Config) *Client {
	timeout := time.Duration(cfg.NVIDIA.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		apiKey:    cfg.NVIDIA.APIKey,
		baseURL:   cfg.NVIDIA.BaseURL,
		llmModel:  cfg.NVIDIA.LLMModel,
		fastModel: cfg.NVIDIA.FastModel,
		codeModel: cfg.NVIDIA.CodingModel,
		http:      &http.Client{Timeout: timeout},
	}
}

func (c *Client) IsConfigured() bool { return c.apiKey != "" }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat interface{}   `json:"response_format,omitempty"`
	Temperature    float64       `json:"temperature"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

func (c *Client) complete(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	body, _ := json.Marshal(chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0.2,
		MaxTokens:      4096,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("nvidia nim request: %w", err)
	}
	defer resp.Body.Close()

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode nvidia response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("nvidia nim error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("nvidia nim: no choices returned")
	}
	return result.Choices[0].Message.Content, nil
}

// ChatJSON calls the primary NIM LLM model with JSON output mode.
func (c *Client) ChatJSON(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.complete(ctx, c.llmModel, systemPrompt, userPrompt)
}

// ChatFastJSON calls the fast NIM model for classification, repair, or short tasks.
func (c *Client) ChatFastJSON(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.complete(ctx, c.fastModel, systemPrompt, userPrompt)
}

// ChatCodingJSON calls the coding NIM model for code-related tasks.
func (c *Client) ChatCodingJSON(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.complete(ctx, c.codeModel, systemPrompt, userPrompt)
}

// GenerateStructured calls NIM and unmarshals into dst.
// Retries once with RepairJSON if the first response is malformed.
func (c *Client) GenerateStructured(ctx context.Context, systemPrompt, userPrompt string, dst interface{}) error {
	raw, err := c.ChatJSON(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		repaired, repairErr := c.RepairJSON(ctx, raw, "")
		if repairErr != nil {
			return fmt.Errorf("json unmarshal failed and repair failed: %w (original: %v)", repairErr, err)
		}
		if err2 := json.Unmarshal([]byte(repaired), dst); err2 != nil {
			return fmt.Errorf("json repair did not fix output: %w", err2)
		}
	}
	return nil
}

// RepairJSON asks the fast NIM model to fix invalid JSON.
func (c *Client) RepairJSON(ctx context.Context, invalidJSON, schemaHint string) (string, error) {
	system := "You are a JSON repair assistant. Return only valid JSON, nothing else. No prose, no markdown fences."
	user := fmt.Sprintf("Fix this malformed JSON and return only the corrected JSON:\n\n%s", invalidJSON)
	if schemaHint != "" {
		user += fmt.Sprintf("\n\nExpected shape hint: %s", schemaHint)
	}
	return c.ChatFastJSON(ctx, system, user)
}

// ProviderStatus returns whether the NVIDIA NIM provider is reachable.
func (c *Client) ProviderStatus() map[string]interface{} {
	return map[string]interface{}{
		"configured": c.IsConfigured(),
		"llmModel":   c.llmModel,
		"fastModel":  c.fastModel,
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "… [truncated]"
}

// Truncate is exported for use by agents.
func Truncate(s string, max int) string { return truncate(s, max) }
