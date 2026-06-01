package handler

import (
	"net/http"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
	"github.com/chetas1208/ActR.AI/apps/api/internal/observability"
)

// Handler is the Vercel function entry point for GET /api/ready
func Handler(w http.ResponseWriter, r *http.Request) {
	if httpx.ApplyCORS(w, r, "*") {
		return
	}

	app, err := observability.GetApp()
	if err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":     false,
			"detail": err.Error(),
		})
		return
	}

	cfg := app.Cfg
	httpStatus := http.StatusOK
	checks := map[string]interface{}{}

	// Core requirements
	if !cfg.Storage.IsConfigured() {
		checks["storage"] = "not_configured"
		httpStatus = http.StatusServiceUnavailable
	} else {
		checks["storage"] = "ok"
	}

	if app.DB == nil {
		checks["db"] = "not_configured"
		httpStatus = http.StatusServiceUnavailable
	} else {
		if cfg.DB.IsInsForge() {
			checks["db"] = "insforge"
		} else if cfg.DB.IsPostgres() {
			checks["db"] = "postgres"
		} else {
			checks["db"] = "memory"
		}
	}

	// Optional providers
	checks["nvidiaLLM"] = cfg.NVIDIA.IsConfigured()
	checks["nvidiaASR"] = cfg.NVIDIASR.IsConfigured()
	checks["daytona"] = cfg.Daytona.IsConfigured()
	checks["rtrvr"] = cfg.Rtrvr.IsConfigured()
	checks["youtube"] = cfg.YouTube.IsConfigured()

	httpx.WriteJSON(w, httpStatus, map[string]interface{}{
		"ok":     httpStatus == http.StatusOK,
		"checks": checks,
	})
}
