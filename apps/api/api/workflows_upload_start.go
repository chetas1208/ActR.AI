package handler

import (
	"net/http"
	"time"

	"github.com/chetas1208/gorube-flow/api/internal/httpx"
	"github.com/chetas1208/gorube-flow/api/internal/models"
	"github.com/chetas1208/gorube-flow/api/internal/observability"
	"github.com/google/uuid"
)

type uploadStartRequest struct {
	JobID     string `json:"jobId"`
	ObjectKey string `json:"objectKey"`
	Title     string `json:"title,omitempty"`
}

type uploadStartResponse struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

// Handler is the Vercel function entry point for POST /api/workflows/upload/start
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

	var req uploadStartRequest
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
	if req.ObjectKey == "" {
		httpx.WriteValidationError(w, "objectKey", "objectKey is required")
		return
	}

	// Verify the object exists in Tigris
	if err := app.Storage.HeadObject(r.Context(), req.ObjectKey); err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "uploaded object not found in storage; complete the Tigris upload first")
		return
	}

	job, err := app.DB.GetWorkflowJob(r.Context(), jobID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "workflow job not found")
		return
	}

	// Update job with source key
	if err := app.DB.UpdateWorkflowJobSource(r.Context(), jobID, req.ObjectKey); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update job")
		return
	}

	// Update title if provided
	if req.Title != "" && (job.Title == nil || *job.Title == "") {
		job.Title = &req.Title
	}

	// Create or update video record
	video, _ := app.DB.GetVideoByJobID(r.Context(), jobID)
	now := time.Now()
	if video == nil {
		title := ""
		if job.Title != nil {
			title = *job.Title
		}
		video = &models.Video{
			ID:             uuid.New(),
			JobID:          jobID,
			Title:          title,
			TigrisVideoKey: &req.ObjectKey,
			Status:         "uploaded",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_ = app.DB.CreateVideo(r.Context(), video)
	} else {
		video.TigrisVideoKey = &req.ObjectKey
		video.Status = "uploaded"
		video.UpdatedAt = now
		_ = app.DB.UpdateVideo(r.Context(), video)
	}

	// Advance to source_ready and create initial workflow steps
	_ = app.DB.UpdateWorkflowJobStatus(r.Context(), jobID, models.StatusSourceReady, "source_ready", 10)

	// Start first bounded workflow step immediately
	result, _ := app.Workflow.Advance(r.Context(), jobID)
	status := models.StatusSourceReady
	if result != nil {
		status = result.NextStatus
	}

	httpx.WriteJSON(w, http.StatusOK, uploadStartResponse{
		JobID:  jobID.String(),
		Status: status,
	})
}
