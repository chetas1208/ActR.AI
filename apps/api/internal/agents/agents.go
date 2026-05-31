package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chetas1208/gorube-flow/api/internal/config"
	"github.com/chetas1208/gorube-flow/api/internal/models"
)

// Client calls an OpenAI-compatible chat completions API.
type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

func NewClient(cfg *config.Config) *Client {
	timeout := time.Duration(cfg.AI.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		apiKey:  cfg.AI.OpenAIKey,
		baseURL: cfg.AI.OpenAIBaseURL,
		model:   cfg.AI.Model,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat interface{}   `json:"response_format,omitempty"`
	Temperature    float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body, _ := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0.3,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode openai response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("openai error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in openai response")
	}
	return result.Choices[0].Message.Content, nil
}

func (c *Client) completeWithRepair(ctx context.Context, system, user string, dst interface{}) error {
	raw, err := c.complete(ctx, system, user)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		repairPrompt := fmt.Sprintf("The following JSON is malformed. Fix it and return only valid JSON:\n\n%s", raw)
		raw2, err2 := c.complete(ctx, "You are a JSON repair assistant. Return only valid JSON, nothing else.", repairPrompt)
		if err2 != nil {
			return fmt.Errorf("json repair request failed: %w", err2)
		}
		if err3 := json.Unmarshal([]byte(raw2), dst); err3 != nil {
			return fmt.Errorf("json repair failed: %w (original: %w)", err3, err)
		}
	}
	return nil
}

// GenerateSummary generates a structured summary from transcript or metadata text.
func (c *Client) GenerateSummary(ctx context.Context, transcriptOrMetadata string) (*models.SummaryJSON, error) {
	system := `You are an expert video content analyst. Analyze the provided transcript or metadata and return a JSON object with exactly these fields:
{
  "title": "string",
  "summary": "2-4 sentence overview",
  "audience": "who this content is for",
  "difficulty": "beginner|intermediate|advanced",
  "keyTopics": ["topic1", "topic2"],
  "chapters": [{"title": "string", "startSeconds": 0, "endSeconds": 60, "summary": "string"}]
}
Return only the JSON object, no other text.`

	user := fmt.Sprintf("Analyze this content:\n\n%s", truncate(transcriptOrMetadata, 12000))

	var result models.SummaryJSON
	if err := c.completeWithRepair(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate summary: %w", err)
	}
	return &result, nil
}

// GenerateActionCards generates actionable items from a summary and transcript.
func (c *Client) GenerateActionCards(ctx context.Context, summary, transcript string) (*models.ActionCardsJSON, error) {
	system := `You are an expert at extracting actionable items from video content. Return a JSON object:
{
  "actions": [
    {
      "title": "string",
      "description": "string",
      "type": "checklist|code|research|study_plan|browser_action",
      "timestampSeconds": 0,
      "requiresExecution": true,
      "provider": "daytona|rtrvr|none"
    }
  ]
}
Use "daytona" for code execution tasks, "rtrvr" for research/browser tasks, "none" for informational items.
Return only the JSON object.`

	user := fmt.Sprintf("Summary:\n%s\n\nTranscript excerpt:\n%s", truncate(summary, 3000), truncate(transcript, 8000))

	var result models.ActionCardsJSON
	if err := c.completeWithRepair(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate action cards: %w", err)
	}
	return &result, nil
}

// ExtractClaims extracts verifiable factual claims from the content.
func (c *Client) ExtractClaims(ctx context.Context, summary, transcript string) (*models.ClaimsJSON, error) {
	system := `You are a fact-checking assistant. Extract verifiable factual claims from this content. Return a JSON object:
{
  "claims": [
    {
      "text": "the specific claim text",
      "timestampSeconds": 0,
      "needsVerification": true,
      "searchQuery": "search query to research this claim"
    }
  ]
}
Focus on statistics, product claims, technical assertions, and pricing information.
Return only the JSON object.`

	user := fmt.Sprintf("Summary:\n%s\n\nTranscript excerpt:\n%s", truncate(summary, 3000), truncate(transcript, 8000))

	var result models.ClaimsJSON
	if err := c.completeWithRepair(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("extract claims: %w", err)
	}
	return &result, nil
}

// GenerateRtrvrResearchTask builds a structured Rtrvr browser research task from claims or an action card.
func (c *Client) GenerateRtrvrResearchTask(ctx context.Context, content string) (*models.RtrvrResearchTaskJSON, error) {
	system := `You are a research task planner. Given content, create a structured browser research task for Rtrvr. Return a JSON object:
{
  "task": "clear task description for the browser agent",
  "targetUrls": ["optional list of URLs to visit"],
  "questions": ["specific questions to answer"],
  "expectedOutput": {
    "sources": ["what source types are expected"],
    "claimAssessments": ["which claims need assessment"],
    "notes": ["any special instructions"]
  }
}
Return only the JSON object.`

	user := fmt.Sprintf("Create a research task for:\n\n%s", truncate(content, 4000))

	var result models.RtrvrResearchTaskJSON
	if err := c.completeWithRepair(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate rtrvr task: %w", err)
	}
	return &result, nil
}

// GenerateExecutionPlan creates a safe code execution plan for a code action card.
func (c *Client) GenerateExecutionPlan(ctx context.Context, actionCard models.ActionCard) (map[string]interface{}, error) {
	system := `You are a code execution planner. Given an action card, produce a safe execution plan as JSON:
{
  "language": "python|javascript|bash|go",
  "code": "the code to execute",
  "description": "what this code does",
  "expectedOutput": "what the output should look like",
  "sandboxRequirements": ["list of packages needed"]
}
Return only the JSON object.`

	user := fmt.Sprintf("Action: %s\nDescription: %s", actionCard.Title, actionCard.Description)

	var result map[string]interface{}
	if err := c.completeWithRepair(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate execution plan: %w", err)
	}
	return result, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "… [truncated]"
}

// sanitizeProviderName ensures provider names do not contain disallowed values.
func sanitizeProviderName(p string) string {
	p = strings.ToLower(strings.TrimSpace(p))
	switch p {
	case "daytona", "rtrvr", "none":
		return p
	}
	return "none"
}

// NormalizeActionCardProvider validates the provider field of a generated action card.
func NormalizeActionCardProvider(provider string) string {
	return sanitizeProviderName(provider)
}
