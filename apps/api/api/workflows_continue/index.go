package handler

import (
	"net/http"
	"strings"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
	"github.com/chetas1208/ActR.AI/apps/api/internal/observability"
)

// Handler is the Vercel function entry point for POST /api/workflows/{jobId}/continue
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

	// Extract jobId from path: /api/workflows/{jobId}/continue
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var jobIDStr string
	for i, p := range parts {
		if p == "workflows" && i+1 < len(parts) {
			jobIDStr = parts[i+1]
			break
		}
	}

	jobID, ok := httpx.ValidateUUID(w, jobIDStr)
	if !ok {
		return
	}

	result, err := app.Workflow.Advance(r.Context(), jobID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	job, _ := app.DB.GetWorkflowJob(r.Context(), jobID)

	resp := map[string]interface{}{
		"jobId":  jobID.String(),
		"status": result.NextStatus,
	}
	if result.OutputKey != "" {
		resp["outputKey"] = result.OutputKey
	}
	if job != nil {
		resp["progress"] = job.Progress
		resp["requiresUserInput"] = job.RequiresUserInput
		if job.RequiredInputType != nil {
			resp["requiredInputType"] = *job.RequiredInputType
		}
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}
