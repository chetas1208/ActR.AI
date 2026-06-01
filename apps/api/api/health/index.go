package handler

import (
	"net/http"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
)

// Handler is the Vercel function entry point for GET /api/health
func Handler(w http.ResponseWriter, r *http.Request) {
	if httpx.ApplyCORS(w, r, "*") {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"service": "actr-ai-api",
	})
}
