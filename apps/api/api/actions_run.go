package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/chetas1208/gorube-flow/api/internal/httpx"
	"github.com/chetas1208/gorube-flow/api/internal/models"
	"github.com/chetas1208/gorube-flow/api/internal/observability"
	"github.com/google/uuid"
)

type actionRunRequest struct {
	Mode string `json:"mode"` // "daytona"
}

// Handler is the Vercel function entry point for POST /api/actions/{actionCardId}/run
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

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var actionCardIDStr string
	for i, p := range parts {
		if p == "actions" && i+1 < len(parts) {
			actionCardIDStr = parts[i+1]
			break
		}
	}
	actionCardID, ok := httpx.ValidateUUID(w, actionCardIDStr)
	if !ok {
		return
	}

	var req actionRunRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Mode == "" {
		req.Mode = "daytona"
	}

	if req.Mode != "daytona" {
		httpx.WriteError(w, http.StatusBadRequest, "only mode 'daytona' is supported for code execution")
		return
	}

	if !app.Daytona.IsConfigured() {
		httpx.ProviderNotConfiguredError(w, "Daytona", "DAYTONA_API_KEY")
		return
	}

	runID := uuid.New()
	now := time.Now()
	run := &models.ExecutionRun{
		ID:           runID,
		ActionCardID: actionCardID,
		Provider:     "daytona",
		Status:       models.RunStatusPending,
		CreatedAt:    now,
	}
	if err := app.DB.CreateExecutionRun(r.Context(), run); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create execution run")
		return
	}

	result, execErr := app.Daytona.RunCode(r.Context(), "python", "print('Gorube Flow: Daytona execution ready')")
	if execErr != nil {
		run.Status = models.RunStatusFailed
		_ = app.DB.UpdateExecutionRun(r.Context(), run)
		httpx.WriteError(w, http.StatusInternalServerError, "execution failed: "+execErr.Error())
		return
	}

	exitCode := result.ExitCode
	run.ExitCode = &exitCode
	run.Status = models.RunStatusCompleted
	logsKey, _ := app.Daytona.StoreExecutionLogs(r.Context(), actionCardID.String(), result)
	if logsKey != "" {
		run.LogsKey = &logsKey
	}
	_ = app.DB.UpdateExecutionRun(r.Context(), run)

	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"runId":    runID.String(),
		"status":   models.RunStatusCompleted,
		"exitCode": exitCode,
		"logsKey":  logsKey,
	})
}
