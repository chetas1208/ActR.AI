package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/chetas1208/gorube-flow/api/internal/httpx"
	"github.com/chetas1208/gorube-flow/api/internal/models"
	"github.com/chetas1208/gorube-flow/api/internal/observability"
	"github.com/google/uuid"
)

// Handler is the Vercel function entry point for POST /api/webhooks/tigris
func Handler(w http.ResponseWriter, r *http.Request) {
	app, err := observability.GetApp()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "service unavailable")
		return
	}
	if !httpx.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Verify webhook secret if configured
	if app.Cfg.Webhook.Secret != "" {
		secret := r.Header.Get("X-Webhook-Secret")
		if secret != app.Cfg.Webhook.Secret {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid webhook secret")
			return
		}
	}

	// Parse payload defensively
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		// Unknown payload shape – log and return 200 to avoid Tigris retries
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored", "reason": "unparseable payload"})
		return
	}

	// Extract object key from common Tigris/S3 webhook shapes
	objectKey := extractObjectKey(payload)
	if objectKey == "" {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored", "reason": "no object key"})
		return
	}

	// Match keys like videos/{jobId}/source/*
	if !strings.HasPrefix(objectKey, "videos/") {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored", "reason": "not a source object"})
		return
	}

	parts := strings.SplitN(objectKey, "/", 4)
	if len(parts) < 4 || parts[2] != "source" {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	jobIDStr := parts[1]
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored", "reason": "invalid job id"})
		return
	}

	// Update job source and advance workflow
	job, err := app.DB.GetWorkflowJob(r.Context(), jobID)
	if err != nil {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored", "reason": "job not found"})
		return
	}

	if job.Status == models.StatusCreated || job.Status == models.StatusQueued {
		_ = app.DB.UpdateWorkflowJobSource(r.Context(), jobID, objectKey)
		_ = app.DB.UpdateWorkflowJobStatus(r.Context(), jobID, models.StatusSourceReady, "source_ready", 10)
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "jobId": jobID.String()})
}

func extractObjectKey(payload map[string]interface{}) string {
	// Try common shapes
	if records, ok := payload["Records"].([]interface{}); ok && len(records) > 0 {
		if rec, ok := records[0].(map[string]interface{}); ok {
			if s3, ok := rec["s3"].(map[string]interface{}); ok {
				if obj, ok := s3["object"].(map[string]interface{}); ok {
					if key, ok := obj["key"].(string); ok {
						return key
					}
				}
			}
		}
	}
	if key, ok := payload["key"].(string); ok {
		return key
	}
	if key, ok := payload["objectKey"].(string); ok {
		return key
	}
	return ""
}
