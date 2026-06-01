package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/chetas1208/ActR.AI/apps/api/internal/observability"
	"github.com/chetas1208/ActR.AI/apps/api/internal/storage"
	"github.com/chetas1208/ActR.AI/apps/api/internal/validation"
	"github.com/google/uuid"
)

type directFileStartRequest struct {
	FileURL               string `json:"fileUrl"`
	Title                 string `json:"title,omitempty"`
	SourceRightsConfirmed bool   `json:"sourceRightsConfirmed"`
}

type directFileStartResponse struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

// Handler is the Vercel function entry point for POST /api/workflows/direct-file/start
func Handler(w http.ResponseWriter, r *http.Request) {
	app, err := observability.GetApp()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "service unavailable")
		return
	}
	if httpx.ApplyCORS(w, r, app.Cfg.App.CORSAllowedOrigins) {
		return
	}
	if !httpx.RequireMethod(w, r, http.MethodPost) {
		return
	}

	var req directFileStartRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !req.SourceRightsConfirmed {
		httpx.WriteValidationError(w, "sourceRightsConfirmed", "you must confirm you have rights to process this file")
		return
	}
	if req.FileURL == "" {
		httpx.WriteValidationError(w, "fileUrl", "fileUrl is required")
		return
	}

	urlInfo, err := validation.ValidateDirectFileURL(r.Context(), req.FileURL, app.Cfg.Upload.MaxDirectFileBytes)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, fmt.Sprintf("URL validation failed: %s", err.Error()))
		return
	}

	jobID := uuid.New()
	now := time.Now()
	title := req.Title
	if title == "" {
		title = "Direct File " + jobID.String()[:8]
	}
	fileURL := req.FileURL
	job := &models.WorkflowJob{
		ID:           jobID,
		SourceType:   models.SourceTypeAuthorizedDirect,
		SourceRights: models.SourceRightsAuthorizedDirect,
		SourceURL:    &fileURL,
		Title:        &title,
		Status:       models.StatusCreated,
		Progress:     0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := app.DB.CreateWorkflowJob(r.Context(), job); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create workflow job")
		return
	}

	ext := strings.ToLower(filepath.Ext(urlInfo.FinalURL))
	if ext == "" {
		ext = extensionFromContentType(urlInfo.ContentType)
	}
	objectKey := fmt.Sprintf("videos/%s/source/original%s", jobID, ext)

	if err := streamDownloadToTigris(r.Context(), urlInfo.FinalURL, objectKey, urlInfo.ContentType, app.Storage); err != nil {
		_ = app.DB.UpdateWorkflowJobError(r.Context(), jobID, "download failed: "+err.Error())
		httpx.WriteError(w, http.StatusBadGateway, "failed to download file: "+err.Error())
		return
	}

	_ = app.DB.UpdateWorkflowJobSource(r.Context(), jobID, objectKey)

	videoRecord := &models.Video{
		ID:             uuid.New(),
		JobID:          jobID,
		Title:          title,
		TigrisVideoKey: &objectKey,
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	_ = app.DB.CreateVideo(r.Context(), videoRecord)
	_ = app.DB.UpdateWorkflowJobStatus(r.Context(), jobID, models.StatusSourceReady, "source_ready", 10)

	result, _ := app.Workflow.Advance(r.Context(), jobID)
	status := models.StatusSourceReady
	if result != nil {
		status = result.NextStatus
	}

	httpx.WriteJSON(w, http.StatusCreated, directFileStartResponse{
		JobID:  jobID.String(),
		Status: status,
	})
}

func streamDownloadToTigris(ctx context.Context, fileURL, objectKey, contentType string, store *storage.Client) error {
	downloadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(downloadCtx, http.MethodGet, fileURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	ct := contentType
	if ct == "" {
		ct = resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/octet-stream"
		}
	}
	return store.StreamUpload(ctx, objectKey, resp.Body, ct)
}

func extensionFromContentType(ct string) string {
	ct = strings.ToLower(strings.SplitN(ct, ";", 2)[0])
	ct = strings.TrimSpace(ct)
	switch ct {
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/webm":
		return ".webm"
	case "audio/mpeg":
		return ".mp3"
	case "audio/mp4":
		return ".m4a"
	case "audio/wav":
		return ".wav"
	case "text/vtt":
		return ".vtt"
	case "text/plain":
		return ".txt"
	case "application/json":
		return ".json"
	default:
		return ".bin"
	}
}
