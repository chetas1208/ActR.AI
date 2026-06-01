package daytona

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/storage"
)

// Client wraps the Daytona sandbox API for safe isolated code execution.
type Client struct {
	apiKey         string
	apiURL         string
	defaultImage   string
	cmdTimeout     time.Duration
	deleteAfterRun bool
	http           *http.Client
	storage        *storage.Client
}

func NewClient(apiKey, apiURL string, cmdTimeout time.Duration, deleteAfterRun bool, store *storage.Client) *Client {
	if apiURL == "" {
		apiURL = "https://app.daytona.io/api"
	}
	if cmdTimeout == 0 {
		cmdTimeout = 60 * time.Second
	}
	return &Client{
		apiKey:         apiKey,
		apiURL:         apiURL,
		defaultImage:   "ubuntu:22.04",
		cmdTimeout:     cmdTimeout,
		deleteAfterRun: deleteAfterRun,
		http:           &http.Client{Timeout: cmdTimeout + 30*time.Second},
		storage:        store,
	}
}

func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

type Sandbox struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ExecutionResult struct {
	SandboxID string `json:"sandboxId"`
	ExitCode  int    `json:"exitCode"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	DurationMs int   `json:"durationMs"`
}

func (c *Client) authHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}

// RunCode creates a sandbox, runs code, captures output, deletes sandbox.
func (c *Client) RunCode(ctx context.Context, language, code string) (*ExecutionResult, error) {
	sandbox, err := c.CreateSandbox(ctx, language)
	if err != nil {
		return nil, fmt.Errorf("create sandbox: %w", err)
	}

	cleanup := c.deleteAfterRun
	defer func() {
		if cleanup {
			_ = c.DeleteSandbox(context.Background(), sandbox.ID)
		}
	}()

	result, err := c.RunCommand(ctx, sandbox.ID, language, code)
	if err != nil {
		return nil, fmt.Errorf("run command: %w", err)
	}
	result.SandboxID = sandbox.ID
	return result, nil
}

// CreateSandbox creates a new Daytona execution sandbox.
func (c *Client) CreateSandbox(ctx context.Context, language string) (*Sandbox, error) {
	image := c.languageToImage(language)
	body, _ := json.Marshal(map[string]interface{}{
		"image":   image,
		"labels":  map[string]string{"app": "actr-ai"},
		"timeout": int(c.cmdTimeout.Seconds()),
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/v1/sandboxes", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.authHeader(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("daytona create sandbox: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("daytona create sandbox returned HTTP %d", resp.StatusCode)
	}

	var sandbox Sandbox
	if err := json.NewDecoder(resp.Body).Decode(&sandbox); err != nil {
		return nil, fmt.Errorf("decode sandbox response: %w", err)
	}
	return &sandbox, nil
}

// RunCommand executes code inside an existing sandbox.
func (c *Client) RunCommand(ctx context.Context, sandboxID, language, code string) (*ExecutionResult, error) {
	cmd := c.buildCommand(language, code)
	body, _ := json.Marshal(map[string]interface{}{
		"command": cmd,
		"timeout": int(c.cmdTimeout.Seconds()),
	})

	url := fmt.Sprintf("%s/v1/sandboxes/%s/exec", c.apiURL, sandboxID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.authHeader(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("daytona exec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daytona exec returned HTTP %d", resp.StatusCode)
	}

	var result ExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode exec response: %w", err)
	}
	return &result, nil
}

// DeleteSandbox terminates and removes a sandbox.
func (c *Client) DeleteSandbox(ctx context.Context, sandboxID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.apiURL+"/v1/sandboxes/"+sandboxID, nil)
	if err != nil {
		return err
	}
	c.authHeader(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// StoreExecutionLogs persists execution results to Tigris.
func (c *Client) StoreExecutionLogs(ctx context.Context, actionCardID string, result *ExecutionResult) (string, error) {
	key := fmt.Sprintf("videos/%s/daytona/execution_logs.json", actionCardID)
	if err := c.storage.PutJSON(ctx, key, result); err != nil {
		return "", err
	}
	return key, nil
}

func (c *Client) languageToImage(lang string) string {
	switch lang {
	case "python":
		return "python:3.12-slim"
	case "javascript", "node":
		return "node:20-slim"
	case "go":
		return "golang:1.22-slim"
	default:
		return c.defaultImage
	}
}

func (c *Client) buildCommand(language, code string) string {
	switch language {
	case "python":
		return fmt.Sprintf("python3 -c %q", code)
	case "javascript", "node":
		return fmt.Sprintf("node -e %q", code)
	case "go":
		return fmt.Sprintf("echo %q > /tmp/main.go && go run /tmp/main.go", code)
	default:
		return fmt.Sprintf("bash -c %q", code)
	}
}
