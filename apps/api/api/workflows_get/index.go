package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
	"github.com/chetas1208/ActR.AI/apps/api/internal/observability"
	"github.com/chetas1208/ActR.AI/apps/api/internal/storage"
)

// Handler is the Vercel function entry point for GET /api/workflows/{jobId}
func Handler(w http.ResponseWriter, r *http.Request) {
	app, err := observability.GetApp()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "service unavailable")
		return
	}
	if httpx.ApplyCORS(w, r, app.Cfg.App.CORSAllowedOrigins) {
		return
	}
	if !httpx.RequireMethod(w, r, http.MethodGet) {
		return
	}

	// Extract jobId from path: /api/workflows/{jobId}
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

	details, err := app.DB.ListWorkflowDetails(r.Context(), jobID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "workflow job not found")
		return
	}

	// Attach signed artifact URLs for stored objects
	keys := storage.ObjectKeys(jobID.String())
	artifactURLs := make(map[string]string)
	for name, key := range keys {
		if err := app.Storage.HeadObject(r.Context(), key); err == nil {
			url, _ := app.Storage.SignedObjectURL(r.Context(), key, 2*time.Hour)
			if url != "" {
				artifactURLs[name] = url
			}
		}
	}

	// Add signed playback URL for uploaded video
	if details.Video != nil && details.Video.TigrisVideoKey != nil {
		url, _ := app.Storage.SignedObjectURL(r.Context(), *details.Video.TigrisVideoKey, 2*time.Hour)
		if url != "" {
			artifactURLs["playback"] = url
		}
	}

	details.ArtifactURLs = artifactURLs

	httpx.WriteJSON(w, http.StatusOK, details)
}
