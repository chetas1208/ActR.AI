// Package db — MemoryRepository is for LOCAL DEVELOPMENT ONLY. Never use in production.
package db

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/google/uuid"
)

type MemoryRepository struct {
	mu          sync.RWMutex
	jobs        map[uuid.UUID]*models.WorkflowJob
	videos      map[uuid.UUID]*models.Video
	steps       map[uuid.UUID][]models.WorkflowStep
	actionCards map[uuid.UUID][]models.ActionCard
	claims      map[uuid.UUID][]models.Claim
	executions  map[uuid.UUID][]models.ExecutionRun
	browserRuns map[uuid.UUID][]models.BrowserRun // keyed by jobID
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		jobs:        make(map[uuid.UUID]*models.WorkflowJob),
		videos:      make(map[uuid.UUID]*models.Video),
		steps:       make(map[uuid.UUID][]models.WorkflowStep),
		actionCards: make(map[uuid.UUID][]models.ActionCard),
		claims:      make(map[uuid.UUID][]models.Claim),
		executions:  make(map[uuid.UUID][]models.ExecutionRun),
		browserRuns: make(map[uuid.UUID][]models.BrowserRun),
	}
}

func (r *MemoryRepository) CreateWorkflowJob(_ context.Context, job *models.WorkflowJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *job
	r.jobs[job.ID] = &cp
	return nil
}
func (r *MemoryRepository) GetWorkflowJob(_ context.Context, id uuid.UUID) (*models.WorkflowJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	j, ok := r.jobs[id]
	if !ok {
		return nil, fmt.Errorf("workflow job not found: %s", id)
	}
	cp := *j
	return &cp, nil
}
func (r *MemoryRepository) UpdateWorkflowJobStatus(_ context.Context, id uuid.UUID, status, currentStep string, progress int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	j.Status = status
	j.CurrentStep = &currentStep
	j.Progress = progress
	j.UpdatedAt = time.Now()
	return nil
}
func (r *MemoryRepository) UpdateWorkflowJobUserInputRequired(_ context.Context, id uuid.UUID, required bool, inputType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	j.RequiresUserInput = required
	if inputType == "" {
		j.RequiredInputType = nil
	} else {
		j.RequiredInputType = &inputType
	}
	if required {
		j.Status = models.StatusWaitingForUserInput
	}
	j.UpdatedAt = time.Now()
	return nil
}
func (r *MemoryRepository) UpdateWorkflowJobSource(_ context.Context, id uuid.UUID, tigrisKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	j.SourceTigrisKey = &tigrisKey
	j.UpdatedAt = time.Now()
	return nil
}
func (r *MemoryRepository) UpdateWorkflowJobTranscript(_ context.Context, id uuid.UUID, transcriptKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	j.TranscriptTigrisKey = &transcriptKey
	j.UpdatedAt = time.Now()
	return nil
}
func (r *MemoryRepository) UpdateWorkflowJobError(_ context.Context, id uuid.UUID, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return fmt.Errorf("not found")
	}
	j.Status = models.StatusFailed
	j.ErrorMessage = &errMsg
	j.UpdatedAt = time.Now()
	return nil
}
func (r *MemoryRepository) ListRecentJobs(_ context.Context, limit int) ([]models.WorkflowJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var jobs []models.WorkflowJob
	for _, j := range r.jobs {
		jobs = append(jobs, *j)
	}
	sort.Slice(jobs, func(i, k int) bool { return jobs[i].CreatedAt.After(jobs[k].CreatedAt) })
	if len(jobs) > limit {
		jobs = jobs[:limit]
	}
	return jobs, nil
}
func (r *MemoryRepository) UpsertWorkflowStep(_ context.Context, step *models.WorkflowStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	steps := r.steps[step.JobID]
	for i, s := range steps {
		if s.StepName == step.StepName {
			steps[i] = *step
			r.steps[step.JobID] = steps
			return nil
		}
	}
	r.steps[step.JobID] = append(steps, *step)
	return nil
}
func (r *MemoryRepository) ListWorkflowSteps(_ context.Context, jobID uuid.UUID) ([]models.WorkflowStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]models.WorkflowStep, len(r.steps[jobID]))
	copy(cp, r.steps[jobID])
	return cp, nil
}
func (r *MemoryRepository) CreateVideo(_ context.Context, video *models.Video) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *video
	r.videos[video.JobID] = &cp
	return nil
}
func (r *MemoryRepository) GetVideoByJobID(_ context.Context, jobID uuid.UUID) (*models.Video, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v := r.videos[jobID]
	if v == nil {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}
func (r *MemoryRepository) UpdateVideo(_ context.Context, video *models.Video) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *video
	r.videos[video.JobID] = &cp
	return nil
}
func (r *MemoryRepository) CreateActionCards(_ context.Context, cards []models.ActionCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(cards) == 0 {
		return nil
	}
	r.actionCards[cards[0].JobID] = append(r.actionCards[cards[0].JobID], cards...)
	return nil
}
func (r *MemoryRepository) ListActionCards(_ context.Context, jobID uuid.UUID) ([]models.ActionCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]models.ActionCard, len(r.actionCards[jobID]))
	copy(cp, r.actionCards[jobID])
	return cp, nil
}
func (r *MemoryRepository) UpdateActionCard(_ context.Context, card *models.ActionCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, c := range r.actionCards[card.JobID] {
		if c.ID == card.ID {
			r.actionCards[card.JobID][i] = *card
			return nil
		}
	}
	return nil
}
func (r *MemoryRepository) CreateClaims(_ context.Context, claims []models.Claim) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(claims) == 0 {
		return nil
	}
	r.claims[claims[0].JobID] = append(r.claims[claims[0].JobID], claims...)
	return nil
}
func (r *MemoryRepository) ListClaims(_ context.Context, jobID uuid.UUID) ([]models.Claim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]models.Claim, len(r.claims[jobID]))
	copy(cp, r.claims[jobID])
	return cp, nil
}
func (r *MemoryRepository) UpdateClaim(_ context.Context, claim *models.Claim) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, c := range r.claims[claim.JobID] {
		if c.ID == claim.ID {
			r.claims[claim.JobID][i] = *claim
			return nil
		}
	}
	return nil
}
func (r *MemoryRepository) CreateExecutionRun(_ context.Context, run *models.ExecutionRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executions[run.ActionCardID] = append(r.executions[run.ActionCardID], *run)
	return nil
}
func (r *MemoryRepository) UpdateExecutionRun(_ context.Context, run *models.ExecutionRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, ex := range r.executions[run.ActionCardID] {
		if ex.ID == run.ID {
			r.executions[run.ActionCardID][i] = *run
			return nil
		}
	}
	return nil
}
func (r *MemoryRepository) ListExecutionRuns(_ context.Context, actionCardID uuid.UUID) ([]models.ExecutionRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]models.ExecutionRun, len(r.executions[actionCardID]))
	copy(cp, r.executions[actionCardID])
	return cp, nil
}
func (r *MemoryRepository) CreateBrowserRun(_ context.Context, run *models.BrowserRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.browserRuns[run.JobID] = append(r.browserRuns[run.JobID], *run)
	return nil
}
func (r *MemoryRepository) UpdateBrowserRun(_ context.Context, run *models.BrowserRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, br := range r.browserRuns[run.JobID] {
		if br.ID == run.ID {
			r.browserRuns[run.JobID][i] = *run
			return nil
		}
	}
	return nil
}
func (r *MemoryRepository) ListBrowserRuns(_ context.Context, jobID uuid.UUID) ([]models.BrowserRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]models.BrowserRun, len(r.browserRuns[jobID]))
	copy(cp, r.browserRuns[jobID])
	return cp, nil
}
func (r *MemoryRepository) ListWorkflowDetails(ctx context.Context, jobID uuid.UUID) (*models.WorkflowDetails, error) {
	job, err := r.GetWorkflowJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	video, _ := r.GetVideoByJobID(ctx, jobID)
	steps, _ := r.ListWorkflowSteps(ctx, jobID)
	cards, _ := r.ListActionCards(ctx, jobID)
	claims, _ := r.ListClaims(ctx, jobID)
	browserRuns, _ := r.ListBrowserRuns(ctx, jobID)
	var runs []models.ExecutionRun
	for _, c := range cards {
		r2, _ := r.ListExecutionRuns(ctx, c.ID)
		runs = append(runs, r2...)
	}
	return &models.WorkflowDetails{
		Job: job, Video: video, Steps: steps, ActionCards: cards,
		Claims: claims, Executions: runs, BrowserRuns: browserRuns,
	}, nil
}
