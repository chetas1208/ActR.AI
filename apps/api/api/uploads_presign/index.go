package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/chetas1208/ActR.AI/apps/api/internal/observability"
	"github.com/google/uuid"
)

type presignUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	FileSize    int64  `json:"fileSize"`
	Title       string `json:"title,omitempty"`
}

type presignUploadResponse struct {
	JobID     string `json:"jobId"`
	UploadURL string `json:"uploadUrl"`
	ObjectKey string `json:"objectKey"`
}

// Handler is the Vercel function entry point for POST /api/uploads/presign
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

	var req presignUploadRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Filename == "" {
		httpx.WriteValidationError(w, "filename", "filename is required")
		return
	}
	if req.ContentType == "" {
		httpx.WriteValidationError(w, "contentType", "contentType is required")
		return
	}
	if req.FileSize <= 0 {
		httpx.WriteValidationError(w, "fileSize", "fileSize must be positive")
		return
	}
	if app.Cfg.Upload.MaxUploadBytes > 0 && req.FileSize > app.Cfg.Upload.MaxUploadBytes {
		httpx.WriteError(w, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("file exceeds maximum size of %d MB", app.Cfg.Upload.MaxUploadBytes/1024/1024))
		return
	}

	jobID := uuid.New()
	ext := filepath.Ext(req.Filename)
	objectKey := fmt.Sprintf("videos/%s/source/original%s", jobID, ext)

	uploadURL, err := app.Storage.GeneratePresignedUploadURL(r.Context(), objectKey, req.ContentType)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	title := req.Title
	if title == "" {
		title = strings.TrimSuffix(req.Filename, ext)
	}

	now := time.Now()
	job := &models.WorkflowJob{
		ID:           jobID,
		SourceType:   models.SourceTypeUpload,
		SourceRights: models.SourceRightsUploadedByUser,
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

	httpx.WriteJSON(w, http.StatusCreated, presignUploadResponse{
		JobID:     jobID.String(),
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	})
}
