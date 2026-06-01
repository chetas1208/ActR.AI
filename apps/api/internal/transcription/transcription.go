// Package transcription provides a provider interface for audio/video transcription.
// Providers: nvidia_asr, transcript_file, local_whisper (dev-only).
// Vercel Functions must never run local_whisper.
package transcription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/config"
	"github.com/chetas1208/ActR.AI/apps/api/internal/storage"
)

// TranscriptResult is the normalised output from any transcription provider.
type TranscriptResult struct {
	Text     string          `json:"text"`
	Language string          `json:"language,omitempty"`
	Provider string          `json:"provider"`
	Segments []TranscriptSeg `json:"segments,omitempty"`
}

type TranscriptSeg struct {
	StartSeconds float64 `json:"startSeconds"`
	EndSeconds   float64 `json:"endSeconds"`
	Text         string  `json:"text"`
}

// Provider is the interface all transcription backends satisfy.
type Provider interface {
	Transcribe(ctx context.Context, sourceObjectKey string) (*TranscriptResult, error)
	IsConfigured() bool
	Name() string
}

// Manager selects the correct provider based on config.
type Manager struct {
	primary   Provider
	storage   *storage.Client
}

func NewManager(cfg *config.Config, store *storage.Client) *Manager {
	var primary Provider

	switch strings.ToLower(cfg.Transcription.Provider) {
	case "local_whisper":
		primary = &localWhisperProvider{cfg: cfg}
	default:
		// Default to NVIDIA ASR if configured
		primary = &nvidiaASRProvider{cfg: cfg, store: store}
	}

	return &Manager{primary: primary, storage: store}
}

// Transcribe runs transcription using the configured provider.
// Returns (nil, ErrProviderNotConfigured) if no provider is available.
func (m *Manager) Transcribe(ctx context.Context, jobID, sourceObjectKey string) (*TranscriptResult, string, error) {
	if !m.primary.IsConfigured() {
		return nil, "", ErrProviderNotConfigured(m.primary.Name())
	}

	result, err := m.primary.Transcribe(ctx, sourceObjectKey)
	if err != nil {
		return nil, "", fmt.Errorf("transcription (%s): %w", m.primary.Name(), err)
	}

	// Store transcript in Tigris
	outputKey := fmt.Sprintf("videos/%s/transcript/transcript.json", jobID)
	if err := m.storage.PutJSON(ctx, outputKey, result); err != nil {
		return result, "", fmt.Errorf("store transcript: %w", err)
	}

	return result, outputKey, nil
}

func (m *Manager) IsASRConfigured() bool { return m.primary.IsConfigured() }

// ProviderNotConfiguredError is a sentinel for upstream caller handling.
type ProviderNotConfiguredError struct{ Name string }

func (e ProviderNotConfiguredError) Error() string {
	return fmt.Sprintf("transcription provider not configured: %s", e.Name)
}

func ErrProviderNotConfigured(name string) error { return ProviderNotConfiguredError{Name: name} }

// --- NVIDIA ASR provider ---

type nvidiaASRProvider struct {
	cfg   *config.Config
	store *storage.Client
}

func (p *nvidiaASRProvider) Name() string { return "nvidia_asr" }

func (p *nvidiaASRProvider) IsConfigured() bool {
	return p.cfg.NVIDIASR.APIKey != "" && p.cfg.Transcription.Provider == "nvidia"
}

func (p *nvidiaASRProvider) Transcribe(ctx context.Context, sourceObjectKey string) (*TranscriptResult, error) {
	// Retrieve audio/video bytes from Tigris
	audioBytes, err := p.store.GetObjectBytes(ctx, sourceObjectKey)
	if err != nil {
		return nil, fmt.Errorf("get source from tigris: %w", err)
	}

	timeout := time.Duration(p.cfg.NVIDIASR.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// NVIDIA ASR API call — multipart form with audio file
	var body bytes.Buffer
	boundary := "----NVIDIAASRBoundary"
	body.WriteString("--" + boundary + "\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"audio\"; filename=\"audio.mp3\"\r\n")
	body.WriteString("Content-Type: audio/mpeg\r\n\r\n")
	body.Write(audioBytes)
	body.WriteString("\r\n--" + boundary + "--\r\n")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.cfg.NVIDIASR.BaseURL+"/audio/transcriptions",
		&body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	req.Header.Set("Authorization", "Bearer "+p.cfg.NVIDIASR.APIKey)

	// Add model and language as query params
	q := req.URL.Query()
	q.Set("model", p.cfg.NVIDIASR.Model)
	if p.cfg.NVIDIASR.Language != "" {
		q.Set("language", p.cfg.NVIDIASR.Language)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nvidia asr request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nvidia asr returned HTTP %d", resp.StatusCode)
	}

	// NVIDIA ASR returns OpenAI-compatible transcription format
	var result struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Segments []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		} `json:"segments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode nvidia asr response: %w", err)
	}

	tr := &TranscriptResult{
		Text:     result.Text,
		Language: result.Language,
		Provider: "nvidia_asr",
	}
	for _, seg := range result.Segments {
		tr.Segments = append(tr.Segments, TranscriptSeg{
			StartSeconds: seg.Start,
			EndSeconds:   seg.End,
			Text:         seg.Text,
		})
	}

	return tr, nil
}

// --- Transcript file provider (parses uploaded .vtt/.srt/.txt) ---

type transcriptFileProvider struct{}

func (p *transcriptFileProvider) Name() string        { return "transcript_file" }
func (p *transcriptFileProvider) IsConfigured() bool  { return true }

func (p *transcriptFileProvider) Transcribe(_ context.Context, content string) (*TranscriptResult, error) {
	return &TranscriptResult{
		Text:     content,
		Provider: "transcript_file",
	}, nil
}

// ParseTranscriptFile parses a raw transcript file into a TranscriptResult.
func ParseTranscriptFile(content string) *TranscriptResult {
	return &TranscriptResult{
		Text:     content,
		Provider: "transcript_file",
	}
}

// --- Local Whisper provider (dev/self-hosted only) ---

type localWhisperProvider struct {
	cfg *config.Config
}

func (p *localWhisperProvider) Name() string { return "local_whisper" }

func (p *localWhisperProvider) IsConfigured() bool {
	return p.cfg.Transcription.LocalEnabled &&
		p.cfg.Transcription.LocalWhisperServerURL != "" &&
		p.cfg.App.Env != "production"
}

func (p *localWhisperProvider) Transcribe(ctx context.Context, sourceObjectKey string) (*TranscriptResult, error) {
	if p.cfg.App.Env == "production" {
		return nil, fmt.Errorf(
			"local Whisper transcription is not available on Vercel. " +
				"Configure NVIDIA ASR (NVIDIA_ASR_API_KEY) or upload a transcript file",
		)
	}
	if !p.cfg.Transcription.LocalEnabled {
		return nil, ErrProviderNotConfigured("local_whisper")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.cfg.Transcription.LocalWhisperServerURL+"/transcribe",
		strings.NewReader(`{"key":"`+sourceObjectKey+`"}`),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 300 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("local whisper request: %w", err)
	}
	defer resp.Body.Close()

	var result TranscriptResult
	result.Provider = "local_whisper"
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		body, _ := io.ReadAll(resp.Body)
		result.Text = string(body)
	}
	return &result, nil
}
