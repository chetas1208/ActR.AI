package handler

import (
	"net/http"
	"time"

	"github.com/chetas1208/gorube-flow/api/internal/httpx"
	"github.com/chetas1208/gorube-flow/api/internal/models"
	"github.com/chetas1208/gorube-flow/api/internal/observability"
	"github.com/google/uuid"
)

type rtrvrRunRequest struct {
	JobID        string   `json:"jobId"`
	ActionCardID string   `json:"actionCardId,omitempty"`
	Task         string   `json:"task"`
	TargetURLs   []string `json:"targetUrls,omitempty"`
}

// Handler is the Vercel function entry point for POST /api/rtrvr/run
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

	if !app.Rtrvr.IsConfigured() {
		httpx.ProviderNotConfiguredError(w, "Rtrvr", "RTRVR_API_KEY")
		return
	}

	var req rtrvrRunRequest
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
	if req.Task == "" {
		httpx.WriteValidationError(w, "task", "task is required")
		return
	}

	var actionCardID *uuid.UUID
	if req.ActionCardID != "" {
		id, parseErr := uuid.Parse(req.ActionCardID)
		if parseErr == nil {
			actionCardID = &id
		}
	}

	result, outputKey, runErr := app.Rtrvr.RunBrowserAction(r.Context(), jobID.String(), req.Task, req.TargetURLs)

	status := models.RunStatusCompleted
	if runErr != nil {
		status = models.RunStatusFailed
	}

	now := time.Now()
	browserRun := &models.BrowserRun{
		ID:           uuid.New(),
		JobID:        jobID,
		ActionCardID: actionCardID,
		Provider:     "rtrvr",
		Status:       status,
		Task:         req.Task,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if outputKey != "" {
		browserRun.OutputKey = &outputKey
	}
	_ = app.DB.CreateBrowserRun(r.Context(), browserRun)

	if runErr != nil {
		httpx.WriteError(w, http.StatusInternalServerError, runErr.Error())
		return
	}

	resp := map[string]interface{}{
		"browserRunId": browserRun.ID.String(),
		"status":       status,
		"outputKey":    outputKey,
	}
	if result != nil {
		resp["taskId"] = result.TaskID
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}
