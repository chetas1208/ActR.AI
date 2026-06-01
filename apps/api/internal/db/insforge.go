package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/google/uuid"
)

// InsForgeRepository implements Repository using the InsForge HTTP API.
type InsForgeRepository struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewInsForgeRepository(baseURL, apiKey string) *InsForgeRepository {
	return &InsForgeRepository{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type insForgeQuery struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params,omitempty"`
}

type insForgeResult struct {
	Rows         []map[string]interface{} `json:"rows"`
	RowsAffected int64                    `json:"rowsAffected"`
	Error        string                   `json:"error,omitempty"`
}

func (r *InsForgeRepository) query(ctx context.Context, sqlStr string, params ...interface{}) (*insForgeResult, error) {
	body, _ := json.Marshal(insForgeQuery{SQL: sqlStr, Params: params})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL+"/query", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("insforge query: %w", err)
	}
	defer resp.Body.Close()

	var result insForgeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode insforge response: %w", err)
	}
	if result.Error != "" {
		return nil, fmt.Errorf("insforge error: %s", result.Error)
	}
	return &result, nil
}

func (r *InsForgeRepository) CreateWorkflowJob(ctx context.Context, job *models.WorkflowJob) error {
	_, err := r.query(ctx, `
		INSERT INTO workflow_jobs (id,user_id,source_type,source_rights,source_url,source_tigris_key,transcript_tigris_key,title,status,current_step,progress,requires_user_input,required_input_type,error_message,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		job.ID.String(), nilStr(job.UserID), job.SourceType, job.SourceRights, job.SourceURL, job.SourceTigrisKey,
		job.TranscriptTigrisKey, job.Title, job.Status, job.CurrentStep, job.Progress,
		job.RequiresUserInput, job.RequiredInputType, job.ErrorMessage, job.CreatedAt, job.UpdatedAt)
	return err
}

func (r *InsForgeRepository) GetWorkflowJob(ctx context.Context, id uuid.UUID) (*models.WorkflowJob, error) {
	result, err := r.query(ctx, `SELECT * FROM workflow_jobs WHERE id=$1 LIMIT 1`, id.String())
	if err != nil {
		return nil, err
	}
	if len(result.Rows) == 0 {
		return nil, fmt.Errorf("workflow job not found: %s", id)
	}
	return rowToWorkflowJob(result.Rows[0])
}

func (r *InsForgeRepository) UpdateWorkflowJobStatus(ctx context.Context, id uuid.UUID, status, currentStep string, progress int) error {
	_, err := r.query(ctx, `UPDATE workflow_jobs SET status=$1,current_step=$2,progress=$3,updated_at=NOW() WHERE id=$4`,
		status, currentStep, progress, id.String())
	return err
}

func (r *InsForgeRepository) UpdateWorkflowJobUserInputRequired(ctx context.Context, id uuid.UUID, required bool, inputType string) error {
	_, err := r.query(ctx, `UPDATE workflow_jobs
		SET requires_user_input=$1,
		    required_input_type=NULLIF($2, ''),
		    status=CASE WHEN $1 THEN $3 ELSE status END,
		    updated_at=NOW()
		WHERE id=$4`,
		required, inputType, models.StatusWaitingForUserInput, id.String())
	return err
}

func (r *InsForgeRepository) UpdateWorkflowJobSource(ctx context.Context, id uuid.UUID, tigrisKey string) error {
	_, err := r.query(ctx, `UPDATE workflow_jobs SET source_tigris_key=$1,updated_at=NOW() WHERE id=$2`, tigrisKey, id.String())
	return err
}

func (r *InsForgeRepository) UpdateWorkflowJobTranscript(ctx context.Context, id uuid.UUID, transcriptKey string) error {
	_, err := r.query(ctx, `UPDATE workflow_jobs SET transcript_tigris_key=$1,updated_at=NOW() WHERE id=$2`, transcriptKey, id.String())
	return err
}

func (r *InsForgeRepository) UpdateWorkflowJobError(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.query(ctx, `UPDATE workflow_jobs SET status=$1,error_message=$2,updated_at=NOW() WHERE id=$3`,
		models.StatusFailed, errMsg, id.String())
	return err
}

func (r *InsForgeRepository) ListRecentJobs(ctx context.Context, limit int) ([]models.WorkflowJob, error) {
	result, err := r.query(ctx, `SELECT * FROM workflow_jobs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	var jobs []models.WorkflowJob
	for _, row := range result.Rows {
		j, err := rowToWorkflowJob(row)
		if err == nil {
			jobs = append(jobs, *j)
		}
	}
	return jobs, nil
}

func (r *InsForgeRepository) UpsertWorkflowStep(ctx context.Context, step *models.WorkflowStep) error {
	_, err := r.query(ctx, `
		INSERT INTO workflow_steps (id,job_id,step_name,status,started_at,finished_at,output_key,error_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (job_id,step_name) DO UPDATE SET status=EXCLUDED.status,started_at=EXCLUDED.started_at,
		  finished_at=EXCLUDED.finished_at,output_key=EXCLUDED.output_key,error_message=EXCLUDED.error_message`,
		step.ID.String(), step.JobID.String(), step.StepName, step.Status, step.StartedAt, step.FinishedAt, step.OutputKey, step.ErrorMessage)
	return err
}

func (r *InsForgeRepository) ListWorkflowSteps(ctx context.Context, jobID uuid.UUID) ([]models.WorkflowStep, error) {
	result, err := r.query(ctx, `SELECT * FROM workflow_steps WHERE job_id=$1 ORDER BY started_at ASC NULLS LAST`, jobID.String())
	if err != nil {
		return nil, err
	}
	var steps []models.WorkflowStep
	for _, row := range result.Rows {
		steps = append(steps, rowToWorkflowStep(row))
	}
	return steps, nil
}

func (r *InsForgeRepository) CreateVideo(ctx context.Context, video *models.Video) error {
	_, err := r.query(ctx, `
		INSERT INTO videos (id,job_id,title,description,duration_seconds,thumbnail_url,youtube_video_id,youtube_embed_url,tigris_video_key,transcript_key,playback_url,status,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		video.ID.String(), video.JobID.String(), video.Title, video.Description, video.DurationSeconds, video.ThumbnailURL,
		video.YouTubeVideoID, video.YouTubeEmbedURL, video.TigrisVideoKey, video.TranscriptKey,
		video.PlaybackURL, video.Status, video.CreatedAt, video.UpdatedAt)
	return err
}

func (r *InsForgeRepository) GetVideoByJobID(ctx context.Context, jobID uuid.UUID) (*models.Video, error) {
	result, err := r.query(ctx, `SELECT * FROM videos WHERE job_id=$1 LIMIT 1`, jobID.String())
	if err != nil || len(result.Rows) == 0 {
		return nil, nil
	}
	return rowToVideo(result.Rows[0]), nil
}

func (r *InsForgeRepository) UpdateVideo(ctx context.Context, video *models.Video) error {
	_, err := r.query(ctx, `
		UPDATE videos SET title=$1,description=$2,duration_seconds=$3,thumbnail_url=$4,
		  youtube_video_id=$5,youtube_embed_url=$6,tigris_video_key=$7,transcript_key=$8,
		  playback_url=$9,status=$10,updated_at=NOW() WHERE id=$11`,
		video.Title, video.Description, video.DurationSeconds, video.ThumbnailURL,
		video.YouTubeVideoID, video.YouTubeEmbedURL, video.TigrisVideoKey, video.TranscriptKey,
		video.PlaybackURL, video.Status, video.ID.String())
	return err
}

func (r *InsForgeRepository) CreateActionCards(ctx context.Context, cards []models.ActionCard) error {
	for _, c := range cards {
		_, err := r.query(ctx, `
			INSERT INTO action_cards (id,job_id,title,description,action_type,timestamp_seconds,status,provider,requires_execution,output_key,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			c.ID.String(), c.JobID.String(), c.Title, c.Description, c.ActionType, c.TimestampSeconds,
			c.Status, c.Provider, c.RequiresExecution, c.OutputKey, c.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *InsForgeRepository) ListActionCards(ctx context.Context, jobID uuid.UUID) ([]models.ActionCard, error) {
	result, err := r.query(ctx, `SELECT * FROM action_cards WHERE job_id=$1 ORDER BY created_at ASC`, jobID.String())
	if err != nil {
		return nil, err
	}
	var cards []models.ActionCard
	for _, row := range result.Rows {
		cards = append(cards, rowToActionCard(row))
	}
	return cards, nil
}

func (r *InsForgeRepository) UpdateActionCard(ctx context.Context, card *models.ActionCard) error {
	_, err := r.query(ctx, `UPDATE action_cards SET status=$1,output_key=$2 WHERE id=$3`,
		card.Status, card.OutputKey, card.ID.String())
	return err
}

func (r *InsForgeRepository) CreateClaims(ctx context.Context, claims []models.Claim) error {
	for _, c := range claims {
		_, err := r.query(ctx, `
			INSERT INTO claims (id,job_id,claim_text,timestamp_seconds,verification_status,confidence,evidence_key,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			c.ID.String(), c.JobID.String(), c.ClaimText, c.TimestampSeconds, c.VerificationStatus, c.Confidence, c.EvidenceKey, c.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *InsForgeRepository) ListClaims(ctx context.Context, jobID uuid.UUID) ([]models.Claim, error) {
	result, err := r.query(ctx, `SELECT * FROM claims WHERE job_id=$1 ORDER BY created_at ASC`, jobID.String())
	if err != nil {
		return nil, err
	}
	var claims []models.Claim
	for _, row := range result.Rows {
		claims = append(claims, rowToClaim(row))
	}
	return claims, nil
}

func (r *InsForgeRepository) UpdateClaim(ctx context.Context, claim *models.Claim) error {
	_, err := r.query(ctx, `UPDATE claims SET verification_status=$1,confidence=$2,evidence_key=$3 WHERE id=$4`,
		claim.VerificationStatus, claim.Confidence, claim.EvidenceKey, claim.ID.String())
	return err
}

func (r *InsForgeRepository) CreateExecutionRun(ctx context.Context, run *models.ExecutionRun) error {
	_, err := r.query(ctx, `
		INSERT INTO execution_runs (id,action_card_id,provider,status,input_key,output_key,logs_key,exit_code,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		run.ID.String(), run.ActionCardID.String(), run.Provider, run.Status, run.InputKey, run.OutputKey, run.LogsKey, run.ExitCode, run.CreatedAt)
	return err
}

func (r *InsForgeRepository) UpdateExecutionRun(ctx context.Context, run *models.ExecutionRun) error {
	_, err := r.query(ctx, `UPDATE execution_runs SET status=$1,output_key=$2,logs_key=$3,exit_code=$4 WHERE id=$5`,
		run.Status, run.OutputKey, run.LogsKey, run.ExitCode, run.ID.String())
	return err
}

func (r *InsForgeRepository) ListExecutionRuns(ctx context.Context, actionCardID uuid.UUID) ([]models.ExecutionRun, error) {
	result, err := r.query(ctx, `SELECT * FROM execution_runs WHERE action_card_id=$1 ORDER BY created_at ASC`, actionCardID.String())
	if err != nil {
		return nil, err
	}
	var runs []models.ExecutionRun
	for _, row := range result.Rows {
		runs = append(runs, rowToExecutionRun(row))
	}
	return runs, nil
}

func (r *InsForgeRepository) CreateBrowserRun(ctx context.Context, run *models.BrowserRun) error {
	var acID interface{}
	if run.ActionCardID != nil {
		acID = run.ActionCardID.String()
	}
	_, err := r.query(ctx, `
		INSERT INTO browser_runs (id,job_id,action_card_id,provider,status,task,input_key,output_key,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		run.ID.String(), run.JobID.String(), acID, run.Provider, run.Status, run.Task, run.InputKey, run.OutputKey, run.CreatedAt, run.UpdatedAt)
	return err
}

func (r *InsForgeRepository) UpdateBrowserRun(ctx context.Context, run *models.BrowserRun) error {
	_, err := r.query(ctx, `UPDATE browser_runs SET status=$1,output_key=$2,updated_at=NOW() WHERE id=$3`,
		run.Status, run.OutputKey, run.ID.String())
	return err
}

func (r *InsForgeRepository) ListBrowserRuns(ctx context.Context, jobID uuid.UUID) ([]models.BrowserRun, error) {
	result, err := r.query(ctx, `SELECT * FROM browser_runs WHERE job_id=$1 ORDER BY created_at ASC`, jobID.String())
	if err != nil {
		return nil, err
	}
	var runs []models.BrowserRun
	for _, row := range result.Rows {
		runs = append(runs, rowToBrowserRun(row))
	}
	return runs, nil
}

func (r *InsForgeRepository) ListWorkflowDetails(ctx context.Context, jobID uuid.UUID) (*models.WorkflowDetails, error) {
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

// --- row mappers ---

func nilStr(id *uuid.UUID) interface{} {
	if id == nil {
		return nil
	}
	return id.String()
}

func rowToWorkflowJob(row map[string]interface{}) (*models.WorkflowJob, error) {
	j := &models.WorkflowJob{CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if v, ok := row["id"].(string); ok {
		j.ID, _ = uuid.Parse(v)
	}
	if v, ok := row["source_type"].(string); ok {
		j.SourceType = v
	}
	if v, ok := row["source_rights"].(string); ok {
		j.SourceRights = v
	}
	if v, ok := row["status"].(string); ok {
		j.Status = v
	}
	if v, ok := row["progress"].(float64); ok {
		j.Progress = int(v)
	}
	if v, ok := row["requires_user_input"].(bool); ok {
		j.RequiresUserInput = v
	}
	strPtr := func(key string) *string {
		if v, ok := row[key].(string); ok && v != "" {
			s := v
			return &s
		}
		return nil
	}
	j.Title = strPtr("title")
	j.SourceURL = strPtr("source_url")
	j.SourceTigrisKey = strPtr("source_tigris_key")
	j.TranscriptTigrisKey = strPtr("transcript_tigris_key")
	j.CurrentStep = strPtr("current_step")
	j.RequiredInputType = strPtr("required_input_type")
	j.ErrorMessage = strPtr("error_message")
	return j, nil
}

func rowToWorkflowStep(row map[string]interface{}) models.WorkflowStep {
	s := models.WorkflowStep{}
	if v, ok := row["id"].(string); ok {
		s.ID, _ = uuid.Parse(v)
	}
	if v, ok := row["job_id"].(string); ok {
		s.JobID, _ = uuid.Parse(v)
	}
	if v, ok := row["step_name"].(string); ok {
		s.StepName = v
	}
	if v, ok := row["status"].(string); ok {
		s.Status = v
	}
	if v, ok := row["output_key"].(string); ok {
		s.OutputKey = &v
	}
	if v, ok := row["error_message"].(string); ok {
		s.ErrorMessage = &v
	}
	return s
}

func rowToVideo(row map[string]interface{}) *models.Video {
	v := &models.Video{CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if id, ok := row["id"].(string); ok {
		v.ID, _ = uuid.Parse(id)
	}
	if id, ok := row["job_id"].(string); ok {
		v.JobID, _ = uuid.Parse(id)
	}
	if t, ok := row["title"].(string); ok {
		v.Title = t
	}
	if s, ok := row["status"].(string); ok {
		v.Status = s
	}
	strPtr := func(key string) *string {
		if val, ok := row[key].(string); ok && val != "" {
			s := val
			return &s
		}
		return nil
	}
	v.YouTubeVideoID = strPtr("youtube_video_id")
	v.YouTubeEmbedURL = strPtr("youtube_embed_url")
	v.ThumbnailURL = strPtr("thumbnail_url")
	v.TigrisVideoKey = strPtr("tigris_video_key")
	v.TranscriptKey = strPtr("transcript_key")
	v.PlaybackURL = strPtr("playback_url")
	v.Description = strPtr("description")
	return v
}

func rowToActionCard(row map[string]interface{}) models.ActionCard {
	c := models.ActionCard{CreatedAt: time.Now()}
	if v, ok := row["id"].(string); ok {
		c.ID, _ = uuid.Parse(v)
	}
	if v, ok := row["job_id"].(string); ok {
		c.JobID, _ = uuid.Parse(v)
	}
	if v, ok := row["title"].(string); ok {
		c.Title = v
	}
	if v, ok := row["description"].(string); ok {
		c.Description = v
	}
	if v, ok := row["action_type"].(string); ok {
		c.ActionType = v
	}
	if v, ok := row["status"].(string); ok {
		c.Status = v
	}
	if v, ok := row["requires_execution"].(bool); ok {
		c.RequiresExecution = v
	}
	if v, ok := row["provider"].(string); ok {
		c.Provider = &v
	}
	if v, ok := row["output_key"].(string); ok {
		c.OutputKey = &v
	}
	return c
}

func rowToClaim(row map[string]interface{}) models.Claim {
	c := models.Claim{CreatedAt: time.Now()}
	if v, ok := row["id"].(string); ok {
		c.ID, _ = uuid.Parse(v)
	}
	if v, ok := row["job_id"].(string); ok {
		c.JobID, _ = uuid.Parse(v)
	}
	if v, ok := row["claim_text"].(string); ok {
		c.ClaimText = v
	}
	if v, ok := row["verification_status"].(string); ok {
		c.VerificationStatus = v
	}
	if v, ok := row["evidence_key"].(string); ok {
		c.EvidenceKey = &v
	}
	if v, ok := row["confidence"].(float64); ok {
		c.Confidence = &v
	}
	return c
}

func rowToExecutionRun(row map[string]interface{}) models.ExecutionRun {
	r := models.ExecutionRun{CreatedAt: time.Now()}
	if v, ok := row["id"].(string); ok {
		r.ID, _ = uuid.Parse(v)
	}
	if v, ok := row["action_card_id"].(string); ok {
		r.ActionCardID, _ = uuid.Parse(v)
	}
	if v, ok := row["provider"].(string); ok {
		r.Provider = v
	}
	if v, ok := row["status"].(string); ok {
		r.Status = v
	}
	if v, ok := row["output_key"].(string); ok {
		r.OutputKey = &v
	}
	if v, ok := row["logs_key"].(string); ok {
		r.LogsKey = &v
	}
	if v, ok := row["input_key"].(string); ok {
		r.InputKey = &v
	}
	if v, ok := row["exit_code"].(float64); ok {
		i := int(v)
		r.ExitCode = &i
	}
	return r
}

func rowToBrowserRun(row map[string]interface{}) models.BrowserRun {
	r := models.BrowserRun{Provider: "rtrvr", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if v, ok := row["id"].(string); ok {
		r.ID, _ = uuid.Parse(v)
	}
	if v, ok := row["job_id"].(string); ok {
		r.JobID, _ = uuid.Parse(v)
	}
	if v, ok := row["provider"].(string); ok {
		r.Provider = v
	}
	if v, ok := row["status"].(string); ok {
		r.Status = v
	}
	if v, ok := row["task"].(string); ok {
		r.Task = v
	}
	if v, ok := row["output_key"].(string); ok {
		r.OutputKey = &v
	}
	if v, ok := row["input_key"].(string); ok {
		r.InputKey = &v
	}
	if v, ok := row["action_card_id"].(string); ok && v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			r.ActionCardID = &id
		}
	}
	return r
}
