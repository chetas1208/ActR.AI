package observability

import (
	"log"
	"sync"

	"github.com/chetas1208/gorube-flow/api/internal/agents"
	"github.com/chetas1208/gorube-flow/api/internal/config"
	"github.com/chetas1208/gorube-flow/api/internal/daytona"
	"github.com/chetas1208/gorube-flow/api/internal/db"
	"github.com/chetas1208/gorube-flow/api/internal/rtrvr"
	"github.com/chetas1208/gorube-flow/api/internal/storage"
	"github.com/chetas1208/gorube-flow/api/internal/workflow"
	"github.com/chetas1208/gorube-flow/api/internal/youtube"
)

// App holds all singletons shared across Vercel function handlers.
type App struct {
	Cfg      *config.Config
	DB       db.Repository
	Storage  *storage.Client
	Agents   *agents.Client
	Daytona  *daytona.Client
	Rtrvr    *rtrvr.Client
	YouTube  *youtube.Client
	Workflow *workflow.Engine
}

var (
	appOnce   sync.Once
	sharedApp *App
	initErr   error
)

// GetApp returns the shared App singleton.
func GetApp() (*App, error) {
	appOnce.Do(func() {
		cfg := config.Load()
		log.Printf("[gorube] env=%s bucket=%s", cfg.App.Env, cfg.Storage.Bucket)

		repo, err := db.NewFromConfig(cfg)
		if err != nil {
			log.Printf("[gorube] WARNING: DB init failed: %v — falling back to memory store", err)
			repo = db.NewMemoryRepository()
		}

		store, err := storage.New(cfg)
		if err != nil {
			initErr = err
			return
		}

		ai := agents.NewClient(cfg)
		day := daytona.NewClient(cfg.Daytona.APIKey, cfg.Daytona.APIURL, cfg.Daytona.CommandTimeout, cfg.Daytona.DeleteAfterRun, store)
		rtr := rtrvr.NewClient(cfg.Rtrvr.APIKey, cfg.Rtrvr.APIURL, cfg.Rtrvr.TimeoutSeconds, store)
		yt := youtube.NewClient(cfg.YouTube.APIKey)
		wf := workflow.NewEngine(repo, store, ai, rtr)

		sharedApp = &App{
			Cfg:      cfg,
			DB:       repo,
			Storage:  store,
			Agents:   ai,
			Daytona:  day,
			Rtrvr:    rtr,
			YouTube:  yt,
			Workflow: wf,
		}
	})
	return sharedApp, initErr
}
