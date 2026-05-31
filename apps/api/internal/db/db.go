package db

import (
	"context"
	"fmt"

	"github.com/chetas1208/gorube-flow/api/internal/config"
	"github.com/chetas1208/gorube-flow/api/internal/models"
	"github.com/google/uuid"
)

// Repository is the interface all DB adapters must satisfy.
type Repository interface {
	// Workflow jobs
	CreateWorkflowJob(ctx context.Context, job *models.WorkflowJob) error
	GetWorkflowJob(ctx context.Context, id uuid.UUID) (*models.WorkflowJob, error)
	UpdateWorkflowJobStatus(ctx context.Context, id uuid.UUID, status, currentStep string, progress int) error
	UpdateWorkflowJobUserInputRequired(ctx context.Context, id uuid.UUID, required bool, inputType string) error
	UpdateWorkflowJobSource(ctx context.Context, id uuid.UUID, tigrisKey string) error
	UpdateWorkflowJobTranscript(ctx context.Context, id uuid.UUID, transcriptKey string) error
	UpdateWorkflowJobError(ctx context.Context, id uuid.UUID, errMsg string) error
	ListRecentJobs(ctx context.Context, limit int) ([]models.WorkflowJob, error)

	// Workflow steps
	UpsertWorkflowStep(ctx context.Context, step *models.WorkflowStep) error
	ListWorkflowSteps(ctx context.Context, jobID uuid.UUID) ([]models.WorkflowStep, error)

	// Videos
	CreateVideo(ctx context.Context, video *models.Video) error
	GetVideoByJobID(ctx context.Context, jobID uuid.UUID) (*models.Video, error)
	UpdateVideo(ctx context.Context, video *models.Video) error

	// Action cards
	CreateActionCards(ctx context.Context, cards []models.ActionCard) error
	ListActionCards(ctx context.Context, jobID uuid.UUID) ([]models.ActionCard, error)
	UpdateActionCard(ctx context.Context, card *models.ActionCard) error

	// Claims
	CreateClaims(ctx context.Context, claims []models.Claim) error
	ListClaims(ctx context.Context, jobID uuid.UUID) ([]models.Claim, error)
	UpdateClaim(ctx context.Context, claim *models.Claim) error

	// Execution runs (Daytona)
	CreateExecutionRun(ctx context.Context, run *models.ExecutionRun) error
	UpdateExecutionRun(ctx context.Context, run *models.ExecutionRun) error
	ListExecutionRuns(ctx context.Context, actionCardID uuid.UUID) ([]models.ExecutionRun, error)

	// Browser runs (Rtrvr)
	CreateBrowserRun(ctx context.Context, run *models.BrowserRun) error
	UpdateBrowserRun(ctx context.Context, run *models.BrowserRun) error
	ListBrowserRuns(ctx context.Context, jobID uuid.UUID) ([]models.BrowserRun, error)

	// Aggregated
	ListWorkflowDetails(ctx context.Context, jobID uuid.UUID) (*models.WorkflowDetails, error)
}

// NewFromConfig selects the best available DB adapter.
func NewFromConfig(cfg *config.Config) (Repository, error) {
	if cfg.DB.UseMemory {
		return NewMemoryRepository(), nil
	}
	if cfg.DB.IsInsForge() {
		return NewInsForgeRepository(cfg.DB.InsForgeURL, cfg.DB.InsForgeAPIKey), nil
	}
	if cfg.DB.IsPostgres() {
		return NewPostgresRepository(cfg.DB.DatabaseURL)
	}
	if cfg.App.Env == "production" {
		return nil, fmt.Errorf("no database configured for production: set DATABASE_URL or INSFORGE_API_URL + INSFORGE_API_KEY")
	}
	return NewMemoryRepository(), nil
}
