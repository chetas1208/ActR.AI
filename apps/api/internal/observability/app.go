// Package observability bootstraps ActR.AI's shared singletons.
package observability

import (
	"log"
	"sync"

	"github.com/chetas1208/ActR.AI/apps/api/internal/agents"
	"github.com/chetas1208/ActR.AI/apps/api/internal/config"
	"github.com/chetas1208/ActR.AI/apps/api/internal/daytona"
	"github.com/chetas1208/ActR.AI/apps/api/internal/db"
	"github.com/chetas1208/ActR.AI/apps/api/internal/nvidia"
	"github.com/chetas1208/ActR.AI/apps/api/internal/rtrvr"
	"github.com/chetas1208/ActR.AI/apps/api/internal/storage"
	"github.com/chetas1208/ActR.AI/apps/api/internal/transcription"
	"github.com/chetas1208/ActR.AI/apps/api/internal/workflow"
	"github.com/chetas1208/ActR.AI/apps/api/internal/youtube"
)

// App holds all singletons shared across Vercel function handlers.
type App struct {
	Cfg           *config.Config
	DB            db.Repository
	Storage       *storage.Client
	NVIDIA        *nvidia.Client
	Agents        *agents.Client
	Transcription *transcription.Manager
	Daytona       *daytona.Client
	Rtrvr         *rtrvr.Client
	YouTube       *youtube.Client
	Workflow      *workflow.Engine
}

var (
	appOnce   sync.Once
	sharedApp *App
	initErr   error
)

// GetApp returns the shared App singleton, initialising it on first cold-start call.
func GetApp() (*App, error) {
	appOnce.Do(func() {
		cfg := config.Load()
		log.Printf("[actr-ai] env=%s bucket=%s nim_configured=%v",
			cfg.App.Env, cfg.Storage.Bucket, cfg.NVIDIA.IsConfigured())

		repo, err := db.NewFromConfig(cfg)
		if err != nil {
			log.Printf("[actr-ai] WARNING: DB init failed: %v — falling back to memory store", err)
			repo = db.NewMemoryRepository()
		}

		store, err := storage.New(cfg)
		if err != nil {
			initErr = err
			return
		}

		nim   := nvidia.NewClient(cfg)
		ai    := agents.NewClient(nim)
		tx    := transcription.NewManager(cfg, store)
		day   := daytona.NewClient(cfg.Daytona.APIKey, cfg.Daytona.APIURL, cfg.Daytona.CommandTimeout, cfg.Daytona.DeleteAfterRun, store)
		rtr   := rtrvr.NewClient(cfg.Rtrvr.APIKey, cfg.Rtrvr.APIURL, cfg.Rtrvr.TimeoutSeconds, store)
		yt    := youtube.NewClient(cfg.YouTube.APIKey)
		wf    := workflow.NewEngine(repo, store, ai, rtr)

		sharedApp = &App{
			Cfg:           cfg,
			DB:            repo,
			Storage:       store,
			NVIDIA:        nim,
			Agents:        ai,
			Transcription: tx,
			Daytona:       day,
			Rtrvr:         rtr,
			YouTube:       yt,
			Workflow:       wf,
		}
	})
	return sharedApp, initErr
}
