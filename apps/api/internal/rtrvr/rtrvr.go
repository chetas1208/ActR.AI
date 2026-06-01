package rtrvr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/chetas1208/ActR.AI/apps/api/internal/storage"
)

// Client wraps the Rtrvr.ai browser automation API.
// Rtrvr is the primary provider for browser automation, source gathering,
// docs lookup, GitHub README inspection, and claim/source research.
type Client struct {
	apiKey  string
	apiURL  string
	timeout time.Duration
	http    *http.Client
	storage *storage.Client
}

func NewClient(apiKey, apiURL string, timeoutSeconds int, store *storage.Client) *Client {
	if apiURL == "" {
		apiURL = "https://api.rtrvr.ai"
	}
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	return &Client{
		apiKey:  apiKey,
		apiURL:  apiURL,
		timeout: timeout,
		http:    &http.Client{Timeout: timeout + 5*time.Second},
		storage: store,
	}
}

func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

// BrowserTaskInput is the payload sent to start a Rtrvr task.
type BrowserTaskInput struct {
	Task       string            `json:"task"`
	TargetURLs []string          `json:"targetUrls,omitempty"`
	Questions  []string          `json:"questions,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
}

// BrowserTask is a running or completed Rtrvr task.
type BrowserTask struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// BrowserResult holds the output of a completed task.
type BrowserResult struct {
	TaskID  string                 `json:"taskId"`
	Status  string                 `json:"status"`
	Content string                 `json:"content,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type endpointNotFoundError struct {
	Path   string
	Status int
	Body   string
}

func (e *endpointNotFoundError) Error() string {
	msg := fmt.Sprintf("rtrvr endpoint %s returned HTTP %d", e.Path, e.Status)
	if e.Body != "" {
		msg += ": " + e.Body
	}
	return msg
}

// StartBrowserTask starts a new Rtrvr browser automation task.
func (c *Client) StartBrowserTask(ctx context.Context, input BrowserTaskInput) (*BrowserTask, error) {
	body, _ := json.Marshal(input)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/v1/tasks", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rtrvr start task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		bodyStr := strings.TrimSpace(string(respBody))
		if resp.StatusCode == http.StatusNotFound {
			return nil, &endpointNotFoundError{Path: "/v1/tasks", Status: resp.StatusCode, Body: bodyStr}
		}
		if bodyStr != "" {
			return nil, fmt.Errorf("rtrvr returned HTTP %d on task start: %s", resp.StatusCode, bodyStr)
		}
		return nil, fmt.Errorf("rtrvr returned HTTP %d on task start", resp.StatusCode)
	}

	var task BrowserTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("decode rtrvr task: %w", err)
	}
	return &task, nil
}

type agentRequest struct {
	Input    string   `json:"input"`
	URLs     []string `json:"urls,omitempty"`
	Response struct {
		Verbosity string `json:"verbosity,omitempty"`
	} `json:"response,omitempty"`
}

type agentResponse struct {
	Success      bool                   `json:"success"`
	Status       string                 `json:"status"`
	TrajectoryID string                 `json:"trajectoryId"`
	Output       interface{}            `json:"output,omitempty"`
	Result       interface{}            `json:"result,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Message      string                 `json:"message,omitempty"`
}

func (c *Client) runAgentTask(ctx context.Context, input BrowserTaskInput) (*BrowserResult, error) {
	reqBody := agentRequest{
		Input: input.Task,
		URLs:  input.TargetURLs,
	}
	reqBody.Response.Verbosity = "final"

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/agent", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rtrvr agent run: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		bodyStr := strings.TrimSpace(string(respBody))
		if bodyStr != "" {
			return nil, fmt.Errorf("rtrvr returned HTTP %d on /agent: %s", resp.StatusCode, bodyStr)
		}
		return nil, fmt.Errorf("rtrvr returned HTTP %d on /agent", resp.StatusCode)
	}

	var payload agentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode rtrvr /agent response: %w", err)
	}

	taskID := payload.TrajectoryID
	if taskID == "" {
		taskID = fmt.Sprintf("rtrvr-%d", time.Now().UnixNano())
	}

	status := strings.ToLower(payload.Status)
	resultStatus := "completed"
	if status == "failed" || status == "cancelled" || status == "error" || status == "requires_input" {
		resultStatus = "failed"
	}

	var content string
	if s, ok := payload.Output.(string); ok {
		content = s
	}

	data := map[string]interface{}{
		"output":   payload.Output,
		"result":   payload.Result,
		"metadata": payload.Metadata,
	}
	errMsg := payload.Error
	if errMsg == "" {
		errMsg = payload.Message
	}

	return &BrowserResult{
		TaskID:  taskID,
		Status:  resultStatus,
		Content: content,
		Data:    data,
		Error:   errMsg,
	}, nil
}

// GetBrowserTaskResult polls for the result of a running task.
func (c *Client) GetBrowserTaskResult(ctx context.Context, taskID string) (*BrowserResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL+"/v1/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rtrvr get task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rtrvr get task returned HTTP %d", resp.StatusCode)
	}

	var result BrowserResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode rtrvr result: %w", err)
	}
	result.TaskID = taskID
	return &result, nil
}

// RunBrowserAction runs a task end-to-end with polling and stores the result in Tigris.
func (c *Client) RunBrowserAction(ctx context.Context, jobID, task string, targetURLs []string) (*BrowserResult, string, error) {
	input := BrowserTaskInput{
		Task:       task,
		TargetURLs: targetURLs,
	}
	browserTask, err := c.StartBrowserTask(ctx, input)
	if err != nil {
		var nf *endpointNotFoundError
		if errors.As(err, &nf) {
			result, runErr := c.runAgentTask(ctx, input)
			if runErr != nil {
				return nil, "", fmt.Errorf("start browser task: %w", runErr)
			}
			key, _ := c.StoreBrowserResultToTigris(ctx, jobID, result)
			return result, key, nil
		}
		return nil, "", fmt.Errorf("start browser task: %w", err)
	}

	deadline := time.Now().Add(c.timeout - 5*time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		result, err := c.GetBrowserTaskResult(ctx, browserTask.ID)
		if err != nil {
			continue
		}
		if result.Status == "completed" || result.Status == "failed" {
			key, _ := c.StoreBrowserResultToTigris(ctx, jobID, result)
			return result, key, nil
		}
	}

	return nil, "", fmt.Errorf("rtrvr task timed out after %s", c.timeout)
}

// ResearchClaims uses Rtrvr to research and assess factual claims extracted from a transcript.
func (c *Client) ResearchClaims(ctx context.Context, jobID string, claims []models.Claim, claimItems []models.ClaimItem) ([]models.Claim, string, error) {
	if len(claims) == 0 {
		return claims, "", nil
	}

	// Build task description from claim texts
	var questionLines []string
	for i, claim := range claims {
		q := fmt.Sprintf("%d. %s", i+1, claim.ClaimText)
		questionLines = append(questionLines, q)
	}

	task := fmt.Sprintf(
		"Research and assess the following claims. For each claim, find official sources, documentation, or evidence that supports or contradicts it. Provide source URLs and a brief assessment.\n\nClaims:\n%s",
		joinLines(questionLines),
	)

	var questions []string
	for _, item := range claimItems {
		if item.SearchQuery != "" {
			questions = append(questions, item.SearchQuery)
		}
	}

	result, outputKey, err := c.RunBrowserAction(ctx, jobID, task, nil)
	if err != nil {
		// Mark all claims as uncertain on Rtrvr failure
		for i := range claims {
			claims[i].VerificationStatus = models.ClaimStatusUncertain
		}
		return claims, "", fmt.Errorf("rtrvr research failed: %w", err)
	}

	_ = questions

	// Store evidence key on each claim and mark as researched
	for i := range claims {
		claims[i].VerificationStatus = models.ClaimStatusResearched
		if outputKey != "" {
			claims[i].EvidenceKey = &outputKey
		}
		if result.Data != nil {
			if conf, ok := result.Data["confidence"].(float64); ok {
				claims[i].Confidence = &conf
			}
		}
	}

	return claims, outputKey, nil
}

// ResearchActionCard uses Rtrvr to gather sources relevant to an action card.
func (c *Client) ResearchActionCard(ctx context.Context, jobID string, card models.ActionCard) (*BrowserResult, string, error) {
	task := fmt.Sprintf(
		"Find official documentation, source links, and relevant examples for: %s\n\nDetails: %s",
		card.Title, card.Description,
	)
	return c.RunBrowserAction(ctx, jobID, task, nil)
}

// StoreBrowserResultToTigris persists a Rtrvr result artifact to Tigris.
func (c *Client) StoreBrowserResultToTigris(ctx context.Context, jobID string, result *BrowserResult) (string, error) {
	key := fmt.Sprintf("videos/%s/rtrvr/browser_results.json", jobID)
	if err := c.storage.PutJSON(ctx, key, result); err != nil {
		return "", err
	}
	return key, nil
}

func joinLines(lines []string) string {
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}
