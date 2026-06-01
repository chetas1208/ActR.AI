// Package agents implements ActR.AI workflow intelligence using NVIDIA NIM.
// All LLM calls go through internal/nvidia — no OpenAI API keys or vars used.
package agents

import (
	"context"
	"fmt"

	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/chetas1208/ActR.AI/apps/api/internal/nvidia"
)

// Client wraps the NVIDIA NIM client for workflow-specific generation tasks.
type Client struct {
	nim *nvidia.Client
}

func NewClient(nim *nvidia.Client) *Client {
	return &Client{nim: nim}
}

func (c *Client) IsConfigured() bool { return c.nim.IsConfigured() }

// GenerateSummary generates a structured summary from transcript or metadata text.
func (c *Client) GenerateSummary(ctx context.Context, transcriptOrMetadata string) (*models.SummaryJSON, error) {
	system := `You are an expert video content analyst powered by NVIDIA NIM. Analyze the provided transcript or metadata and return a JSON object with exactly this structure:
{
  "title": "string",
  "summary": "2-4 sentence overview",
  "audience": "who this content is for",
  "difficulty": "beginner|intermediate|advanced",
  "keyTopics": ["topic1", "topic2"],
  "chapters": [{"title": "string", "startSeconds": 0, "endSeconds": 60, "summary": "string"}]
}
Return only the JSON object. No markdown, no code fences, no prose.`

	user := fmt.Sprintf("Analyze this content:\n\n%s", nvidia.Truncate(transcriptOrMetadata, 12000))

	var result models.SummaryJSON
	if err := c.nim.GenerateStructured(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate summary: %w", err)
	}
	return &result, nil
}

// GenerateActionCards generates actionable items from a summary and transcript.
func (c *Client) GenerateActionCards(ctx context.Context, summary, transcript string) (*models.ActionCardsJSON, error) {
	system := `You are an expert at extracting actionable items from video content using NVIDIA NIM reasoning. Return a JSON object:
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

	user := fmt.Sprintf("Summary:\n%s\n\nTranscript excerpt:\n%s",
		nvidia.Truncate(summary, 3000), nvidia.Truncate(transcript, 8000))

	var result models.ActionCardsJSON
	if err := c.nim.GenerateStructured(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate action cards: %w", err)
	}
	return &result, nil
}

// ExtractClaims extracts verifiable factual claims.
func (c *Client) ExtractClaims(ctx context.Context, summary, transcript string) (*models.ClaimsJSON, error) {
	system := `You are a fact-checking assistant using NVIDIA NIM. Extract verifiable factual claims from this content. Return a JSON object:
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

	user := fmt.Sprintf("Summary:\n%s\n\nTranscript excerpt:\n%s",
		nvidia.Truncate(summary, 3000), nvidia.Truncate(transcript, 8000))

	var result models.ClaimsJSON
	if err := c.nim.GenerateStructured(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("extract claims: %w", err)
	}
	return &result, nil
}

// GenerateRtrvrResearchTask builds a structured Rtrvr browser research task.
func (c *Client) GenerateRtrvrResearchTask(ctx context.Context, content string) (*models.RtrvrResearchTaskJSON, error) {
	system := `You are a research task planner. Given content, create a structured browser research task for Rtrvr.ai. Return a JSON object:
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

	user := fmt.Sprintf("Create a research task for:\n\n%s", nvidia.Truncate(content, 4000))

	var result models.RtrvrResearchTaskJSON
	if err := c.nim.GenerateStructured(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate rtrvr task: %w", err)
	}
	return &result, nil
}

// GenerateExecutionPlan creates a safe code execution plan for a Daytona action card.
func (c *Client) GenerateExecutionPlan(ctx context.Context, actionCard models.ActionCard) (map[string]interface{}, error) {
	system := `You are a code execution planner using NVIDIA NIM. Given an action card, produce a safe execution plan as JSON:
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
	if err := c.nim.GenerateStructured(ctx, system, user, &result); err != nil {
		return nil, fmt.Errorf("generate execution plan: %w", err)
	}
	return result, nil
}

// NormalizeActionCardProvider validates the provider field from AI output.
func NormalizeActionCardProvider(p string) string {
	switch p {
	case "daytona", "rtrvr", "none":
		return p
	}
	return "none"
}
