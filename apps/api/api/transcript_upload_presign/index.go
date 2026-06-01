package handler

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
	"github.com/chetas1208/ActR.AI/apps/api/internal/observability"
	"github.com/google/uuid"
)

type transcriptPresignRequest struct {
	JobID       string `json:"jobId"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

type transcriptPresignResponse struct {
	UploadURL string `json:"uploadUrl"`
	ObjectKey string `json:"objectKey"`
}

// Handler is the Vercel function entry point for POST /api/uploads/transcript/presign
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

	var req transcriptPresignRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.JobID == "" {
		httpx.WriteValidationError(w, "jobId", "jobId is required")
		return
	}
	jobID, ok := httpx.ValidateUUID(w, req.JobID)
	if !ok {
		return
	}
	if req.Filename == "" {
		httpx.WriteValidationError(w, "filename", "filename is required")
		return
	}

	ext := filepath.Ext(req.Filename)
	objectKey := fmt.Sprintf("videos/%s/source/transcript%s", jobID, ext)
	ct := req.ContentType
	if ct == "" {
		ct = "text/plain"
	}

	uploadURL, err := app.Storage.GeneratePresignedUploadURL(r.Context(), objectKey, ct)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	_ = uuid.UUID{} // keep import
	httpx.WriteJSON(w, http.StatusOK, transcriptPresignResponse{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	})
}
