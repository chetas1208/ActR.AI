package handler

import (
	"net/http"

	"github.com/chetas1208/gorube-flow/api/internal/httpx"
	"github.com/chetas1208/gorube-flow/api/internal/observability"
)

// Handler is the Vercel function entry point for GET /api/providers/status
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
	if !app.Cfg.App.EnableProviderStatus {
		httpx.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	dbStatus := "missing"
	switch {
	case app.Cfg.DB.IsInsForge():
		dbStatus = "insforge"
	case app.Cfg.DB.IsPostgres():
		dbStatus = "postgres"
	case app.Cfg.DB.UseMemory || app.DB != nil:
		dbStatus = "memory"
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"nvidiaLLM": app.Cfg.NVIDIA.IsConfigured(),
		"nvidiaASR": app.Cfg.NVIDIASR.IsConfigured(),
		"tigris":    app.Cfg.Storage.IsConfigured(),
		"daytona":   app.Cfg.Daytona.IsConfigured(),
		"rtrvr":     app.Cfg.Rtrvr.IsConfigured(),
		"youtube":   app.Cfg.YouTube.IsConfigured(),
		"database":  dbStatus,
	})
}

