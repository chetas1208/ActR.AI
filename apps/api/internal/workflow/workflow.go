package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/chetas1208/gorube-flow/api/internal/agents"
	"github.com/chetas1208/gorube-flow/api/internal/db"
	"github.com/chetas1208/gorube-flow/api/internal/models"
	"github.com/chetas1208/gorube-flow/api/internal/rtrvr"
	"github.com/chetas1208/gorube-flow/api/internal/storage"
	"github.com/google/uuid"
)

// Engine orchestrates bounded workflow steps.
type Engine struct {
	db      db.Repository
	storage *storage.Client
	agents  *agents.Client
	rtrvr   *rtrvr.Client
}

func NewEngine(repo db.Repository, store *storage.Client, ai *agents.Client, rtr *rtrvr.Client) *Engine {
	return &Engine{db: repo, storage: store, agents: ai, rtrvr: rtr}
}

// StepResult is returned by each bounded step.
type StepResult struct {
	NextStatus string
	NextStep   string
	Progress   int
	StepName   string
	OutputKey  string
	NeedsInput bool
	InputType  string
}

// Advance runs the next pending bounded step for a job.
func (e *Engine) Advance(ctx context.Context, jobID uuid.UUID) (*StepResult, error) {
	job, err := e.db.GetWorkflowJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}

	switch job.Status {
	case models.StatusSourceReady:
		return e.runTranscriptStep(ctx, job)
	case models.StatusTranscribing:
		return e.runChunkingStep(ctx, job)
	case models.StatusChunking:
		return e.runSummarizingStep(ctx, job)
	case models.StatusSummarizing:
		return e.runExtractActionsStep(ctx, job)
	case models.StatusExtractingActions:
		return e.runExtractClaimsStep(ctx, job)
	case models.StatusExtractingClaims:
		return e.runRtrvrResearchStep(ctx, job)
	case models.StatusResearchingWithRtrvr:
		return e.runFinalizingStep(ctx, job)
	case models.StatusReady, models.StatusFailed, models.StatusWaitingForUserInput:
		return &StepResult{NextStatus: job.Status}, nil
	default:
		return &StepResult{NextStatus: job.Status}, nil
	}
}

func (e *Engine) startStep(ctx context.Context, jobID uuid.UUID, stepName, status string, progress int) error {
	now := time.Now()
	return e.db.UpsertWorkflowStep(ctx, &models.WorkflowStep{
		ID:        uuid.New(),
		JobID:     jobID,
		StepName:  stepName,
		Status:    "running",
		StartedAt: &now,
	})
}

func (e *Engine) completeStep(ctx context.Context, jobID uuid.UUID, stepName, outputKey string) error {
	now := time.Now()
	step := &models.WorkflowStep{
		ID:         uuid.New(),
		JobID:      jobID,
		StepName:   stepName,
		Status:     "completed",
		FinishedAt: &now,
	}
	if outputKey != "" {
		step.OutputKey = &outputKey
	}
	return e.db.UpsertWorkflowStep(ctx, step)
}

func (e *Engine) failStep(ctx context.Context, jobID uuid.UUID, stepName, errMsg string) error {
	now := time.Now()
	step := &models.WorkflowStep{
		ID:           uuid.New(),
		JobID:        jobID,
		StepName:     stepName,
		Status:       "failed",
		FinishedAt:   &now,
		ErrorMessage: &errMsg,
	}
	_ = e.db.UpsertWorkflowStep(ctx, step)
	return e.db.UpdateWorkflowJobError(ctx, jobID, errMsg)
}

func (e *Engine) runTranscriptStep(ctx context.Context, job *models.WorkflowJob) (*StepResult, error) {
	stepName := "prepare_transcript"
	_ = e.startStep(ctx, job.ID, stepName, models.StatusTranscribing, 10)
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusTranscribing, stepName, 10)

	var transcriptText string

	if job.TranscriptTigrisKey != nil {
		var raw map[string]string
		if err := e.storage.GetJSON(ctx, *job.TranscriptTigrisKey, &raw); err == nil {
			transcriptText = raw["text"]
		}
		if transcriptText == "" {
			transcriptText = "[transcript file loaded]"
		}
	} else if job.SourceType == models.SourceTypeYouTube {
		_ = e.db.UpdateWorkflowJobUserInputRequired(ctx, job.ID, true, models.RequiredInputTranscriptAudioOrVideo)
		_ = e.completeStep(ctx, job.ID, stepName, "")
		return &StepResult{
			NextStatus: models.StatusWaitingForUserInput,
			NeedsInput: true,
			InputType:  models.RequiredInputTranscriptAudioOrVideo,
			Progress:   10,
		}, nil
	} else {
		if !e.agents.IsConfigured() {
			_ = e.db.UpdateWorkflowJobUserInputRequired(ctx, job.ID, true, models.RequiredInputTranscript)
			return &StepResult{
				NextStatus: models.StatusWaitingForUserInput,
				NeedsInput: true,
				InputType:  models.RequiredInputTranscript,
				Progress:   10,
			}, nil
		}
		transcriptText = "[transcript pending — media file uploaded]"
	}

	keys := storage.ObjectKeys(job.ID.String())
	outputKey := keys["transcript_json"]
	_ = e.storage.PutJSON(ctx, outputKey, map[string]string{"text": transcriptText})
	_ = e.completeStep(ctx, job.ID, stepName, outputKey)
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusChunking, "chunking", 20)

	return &StepResult{
		NextStatus: models.StatusChunking,
		Progress:   20,
		StepName:   stepName,
		OutputKey:  outputKey,
	}, nil
}

func (e *Engine) runChunkingStep(ctx context.Context, job *models.WorkflowJob) (*StepResult, error) {
	stepName := "chunk_transcript"
	_ = e.startStep(ctx, job.ID, stepName, models.StatusSummarizing, 30)
	_ = e.completeStep(ctx, job.ID, stepName, "")
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusSummarizing, "summarizing", 35)
	return &StepResult{NextStatus: models.StatusSummarizing, Progress: 35}, nil
}

func (e *Engine) runSummarizingStep(ctx context.Context, job *models.WorkflowJob) (*StepResult, error) {
	stepName := "generate_summary"
	_ = e.startStep(ctx, job.ID, stepName, models.StatusSummarizing, 40)

	if !e.agents.IsConfigured() {
		msg := "AI provider not configured: add NIM_API_KEY to enable summarization via NVIDIA NIM."
		_ = e.failStep(ctx, job.ID, stepName, msg)
		return &StepResult{NextStatus: models.StatusFailed}, nil
	}

	keys := storage.ObjectKeys(job.ID.String())
	var transcriptData map[string]string
	_ = e.storage.GetJSON(ctx, keys["transcript_json"], &transcriptData)
	transcriptText := transcriptData["text"]
	if transcriptText == "" {
		transcriptText = "No transcript available."
	}

	summary, err := e.agents.GenerateSummary(ctx, transcriptText)
	if err != nil {
		_ = e.failStep(ctx, job.ID, stepName, err.Error())
		return &StepResult{NextStatus: models.StatusFailed}, nil
	}

	outputKey, _ := e.storage.StoreArtifact(ctx, job.ID.String(), "analysis", "summary.json", summary)
	_ = e.completeStep(ctx, job.ID, stepName, outputKey)
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusExtractingActions, "extracting_actions", 55)

	return &StepResult{
		NextStatus: models.StatusExtractingActions,
		Progress:   55,
		StepName:   stepName,
		OutputKey:  outputKey,
	}, nil
}

func (e *Engine) runExtractActionsStep(ctx context.Context, job *models.WorkflowJob) (*StepResult, error) {
	stepName := "extract_action_cards"
	_ = e.startStep(ctx, job.ID, stepName, models.StatusExtractingActions, 60)

	keys := storage.ObjectKeys(job.ID.String())
	var summary models.SummaryJSON
	_ = e.storage.GetJSON(ctx, keys["summary"], &summary)
	var transcriptData map[string]string
	_ = e.storage.GetJSON(ctx, keys["transcript_json"], &transcriptData)

	actionCardsJSON, err := e.agents.GenerateActionCards(ctx, summary.Summary, transcriptData["text"])
	if err != nil {
		_ = e.failStep(ctx, job.ID, stepName, err.Error())
		return &StepResult{NextStatus: models.StatusFailed}, nil
	}

	outputKey, _ := e.storage.StoreArtifact(ctx, job.ID.String(), "analysis", "action_cards.json", actionCardsJSON)

	var cards []models.ActionCard
	now := time.Now()
	for _, item := range actionCardsJSON.Actions {
		provider := agents.NormalizeActionCardProvider(item.Provider)
		ts := item.TimestampSeconds
		cards = append(cards, models.ActionCard{
			ID:                uuid.New(),
			JobID:             job.ID,
			Title:             item.Title,
			Description:       item.Description,
			ActionType:        item.Type,
			TimestampSeconds:  &ts,
			Status:            models.ActionStatusPending,
			Provider:          &provider,
			RequiresExecution: item.RequiresExecution,
			CreatedAt:         now,
		})
	}
	if len(cards) > 0 {
		_ = e.db.CreateActionCards(ctx, cards)
	}

	_ = e.completeStep(ctx, job.ID, stepName, outputKey)
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusExtractingClaims, "extracting_claims", 70)

	return &StepResult{
		NextStatus: models.StatusExtractingClaims,
		Progress:   70,
		StepName:   stepName,
		OutputKey:  outputKey,
	}, nil
}

func (e *Engine) runExtractClaimsStep(ctx context.Context, job *models.WorkflowJob) (*StepResult, error) {
	stepName := "extract_claims"
	_ = e.startStep(ctx, job.ID, stepName, models.StatusExtractingClaims, 75)

	keys := storage.ObjectKeys(job.ID.String())
	var summary models.SummaryJSON
	_ = e.storage.GetJSON(ctx, keys["summary"], &summary)
	var transcriptData map[string]string
	_ = e.storage.GetJSON(ctx, keys["transcript_json"], &transcriptData)

	claimsJSON, err := e.agents.ExtractClaims(ctx, summary.Summary, transcriptData["text"])
	if err != nil {
		_ = e.failStep(ctx, job.ID, stepName, err.Error())
		return &StepResult{NextStatus: models.StatusFailed}, nil
	}

	outputKey, _ := e.storage.StoreArtifact(ctx, job.ID.String(), "analysis", "claims.json", claimsJSON)

	var claims []models.Claim
	now := time.Now()
	for _, item := range claimsJSON.Claims {
		ts := item.TimestampSeconds
		claims = append(claims, models.Claim{
			ID:                 uuid.New(),
			JobID:              job.ID,
			ClaimText:          item.Text,
			TimestampSeconds:   &ts,
			VerificationStatus: models.ClaimStatusPending,
			CreatedAt:          now,
		})
	}
	if len(claims) > 0 {
		_ = e.db.CreateClaims(ctx, claims)
	}

	_ = e.completeStep(ctx, job.ID, stepName, outputKey)

	nextStatus := models.StatusResearchingWithRtrvr
	if !e.rtrvr.IsConfigured() {
		nextStatus = models.StatusFinalizing
	}
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, nextStatus, nextStatus, 80)

	return &StepResult{
		NextStatus: nextStatus,
		Progress:   80,
		StepName:   stepName,
		OutputKey:  outputKey,
	}, nil
}

func (e *Engine) runRtrvrResearchStep(ctx context.Context, job *models.WorkflowJob) (*StepResult, error) {
	stepName := "rtrvr_research"
	_ = e.startStep(ctx, job.ID, stepName, models.StatusResearchingWithRtrvr, 85)

	if !e.rtrvr.IsConfigured() {
		_ = e.completeStep(ctx, job.ID, stepName, "")
		_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusFinalizing, "finalizing", 90)
		return &StepResult{NextStatus: models.StatusFinalizing, Progress: 90}, nil
	}

	claims, err := e.db.ListClaims(ctx, job.ID)
	if err != nil || len(claims) == 0 {
		_ = e.completeStep(ctx, job.ID, stepName, "")
		_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusFinalizing, "finalizing", 90)
		return &StepResult{NextStatus: models.StatusFinalizing, Progress: 90}, nil
	}

	keys := storage.ObjectKeys(job.ID.String())
	var claimsJSON models.ClaimsJSON
	_ = e.storage.GetJSON(ctx, keys["claims"], &claimsJSON)

	updatedClaims, outputKey, researchErr := e.rtrvr.ResearchClaims(ctx, job.ID.String(), claims, claimsJSON.Claims)
	if researchErr != nil {
		for i := range updatedClaims {
			updatedClaims[i].VerificationStatus = models.ClaimStatusUncertain
		}
	}

	for _, c := range updatedClaims {
		claim := c
		_ = e.db.UpdateClaim(ctx, &claim)
	}

	// Record the browser run
	now := time.Now()
	_ = e.db.CreateBrowserRun(ctx, &models.BrowserRun{
		ID:        uuid.New(),
		JobID:     job.ID,
		Provider:  "rtrvr",
		Status:    models.RunStatusCompleted,
		Task:      "Research and assess extracted claims",
		OutputKey: &outputKey,
		CreatedAt: now,
		UpdatedAt: now,
	})

	_ = e.completeStep(ctx, job.ID, stepName, outputKey)
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusFinalizing, "finalizing", 90)

	return &StepResult{
		NextStatus: models.StatusFinalizing,
		Progress:   90,
		StepName:   stepName,
		OutputKey:  outputKey,
	}, nil
}

func (e *Engine) runFinalizingStep(ctx context.Context, job *models.WorkflowJob) (*StepResult, error) {
	stepName := "finalize"
	_ = e.startStep(ctx, job.ID, stepName, models.StatusFinalizing, 95)

	details, err := e.db.ListWorkflowDetails(ctx, job.ID)
	if err != nil {
		_ = e.failStep(ctx, job.ID, stepName, err.Error())
		return &StepResult{NextStatus: models.StatusFailed}, nil
	}

	export := buildFinalExport(details)
	outputKey, _ := e.storage.StoreArtifact(ctx, job.ID.String(), "exports", "final_workflow.json", export)

	_ = e.completeStep(ctx, job.ID, stepName, outputKey)
	_ = e.db.UpdateWorkflowJobStatus(ctx, job.ID, models.StatusReady, "ready", 100)

	return &StepResult{
		NextStatus: models.StatusReady,
		Progress:   100,
		StepName:   stepName,
		OutputKey:  outputKey,
	}, nil
}

func buildFinalExport(details *models.WorkflowDetails) map[string]interface{} {
	artifactKeys := make(map[string]string)
	for _, s := range details.Steps {
		if s.OutputKey != nil {
			artifactKeys[s.StepName] = *s.OutputKey
		}
	}
	return map[string]interface{}{
		"jobId":                details.Job.ID,
		"sourceType":           details.Job.SourceType,
		"title":                details.Job.Title,
		"actions":              details.ActionCards,
		"claims":               details.Claims,
		"rtrvrResearchResults": details.BrowserRuns,
		"executionRuns":        details.Executions,
		"artifactKeys":         artifactKeys,
		"exportedAt":           time.Now().UTC().Format(time.RFC3339),
	}
}
