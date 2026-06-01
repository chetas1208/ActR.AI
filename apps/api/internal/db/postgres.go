package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/google/uuid"
)

// PostgresRepository implements Repository using Postgres.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(databaseURL string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreateWorkflowJob(ctx context.Context, job *models.WorkflowJob) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO workflow_jobs (id,user_id,source_type,source_rights,source_url,source_tigris_key,transcript_tigris_key,title,status,current_step,progress,requires_user_input,required_input_type,error_message,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		job.ID, job.UserID, job.SourceType, job.SourceRights, job.SourceURL, job.SourceTigrisKey,
		job.TranscriptTigrisKey, job.Title, job.Status, job.CurrentStep, job.Progress,
		job.RequiresUserInput, job.RequiredInputType, job.ErrorMessage, job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetWorkflowJob(ctx context.Context, id uuid.UUID) (*models.WorkflowJob, error) {
	j := &models.WorkflowJob{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id,user_id,source_type,source_rights,source_url,source_tigris_key,transcript_tigris_key,
		       title,status,current_step,progress,requires_user_input,required_input_type,error_message,created_at,updated_at
		FROM workflow_jobs WHERE id=$1`, id).Scan(
		&j.ID, &j.UserID, &j.SourceType, &j.SourceRights, &j.SourceURL, &j.SourceTigrisKey,
		&j.TranscriptTigrisKey, &j.Title, &j.Status, &j.CurrentStep, &j.Progress,
		&j.RequiresUserInput, &j.RequiredInputType, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workflow job not found: %s", id)
	}
	return j, err
}

func (r *PostgresRepository) UpdateWorkflowJobStatus(ctx context.Context, id uuid.UUID, status, currentStep string, progress int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_jobs SET status=$1,current_step=$2,progress=$3,updated_at=NOW() WHERE id=$4`,
		status, currentStep, progress, id)
	return err
}

func (r *PostgresRepository) UpdateWorkflowJobUserInputRequired(ctx context.Context, id uuid.UUID, required bool, inputType string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_jobs
		 SET requires_user_input=$1,
		     required_input_type=NULLIF($2, ''),
		     status=CASE WHEN $1 THEN $3 ELSE status END,
		     updated_at=NOW()
		 WHERE id=$4`,
		required, inputType, models.StatusWaitingForUserInput, id)
	return err
}

func (r *PostgresRepository) UpdateWorkflowJobSource(ctx context.Context, id uuid.UUID, tigrisKey string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_jobs SET source_tigris_key=$1,updated_at=NOW() WHERE id=$2`, tigrisKey, id)
	return err
}

func (r *PostgresRepository) UpdateWorkflowJobTranscript(ctx context.Context, id uuid.UUID, transcriptKey string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_jobs SET transcript_tigris_key=$1,updated_at=NOW() WHERE id=$2`, transcriptKey, id)
	return err
}

func (r *PostgresRepository) UpdateWorkflowJobError(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_jobs SET status=$1,error_message=$2,updated_at=NOW() WHERE id=$3`,
		models.StatusFailed, errMsg, id)
	return err
}

func (r *PostgresRepository) ListRecentJobs(ctx context.Context, limit int) ([]models.WorkflowJob, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,user_id,source_type,source_rights,source_url,source_tigris_key,transcript_tigris_key,
		       title,status,current_step,progress,requires_user_input,required_input_type,error_message,created_at,updated_at
		FROM workflow_jobs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []models.WorkflowJob
	for rows.Next() {
		var j models.WorkflowJob
		rows.Scan(&j.ID, &j.UserID, &j.SourceType, &j.SourceRights, &j.SourceURL, &j.SourceTigrisKey,
			&j.TranscriptTigrisKey, &j.Title, &j.Status, &j.CurrentStep, &j.Progress,
			&j.RequiresUserInput, &j.RequiredInputType, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt)
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *PostgresRepository) UpsertWorkflowStep(ctx context.Context, step *models.WorkflowStep) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO workflow_steps (id,job_id,step_name,status,started_at,finished_at,output_key,error_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (job_id,step_name) DO UPDATE SET
		  status=EXCLUDED.status,started_at=EXCLUDED.started_at,finished_at=EXCLUDED.finished_at,
		  output_key=EXCLUDED.output_key,error_message=EXCLUDED.error_message`,
		step.ID, step.JobID, step.StepName, step.Status, step.StartedAt, step.FinishedAt, step.OutputKey, step.ErrorMessage)
	return err
}

func (r *PostgresRepository) ListWorkflowSteps(ctx context.Context, jobID uuid.UUID) ([]models.WorkflowStep, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,job_id,step_name,status,started_at,finished_at,output_key,error_message
		FROM workflow_steps WHERE job_id=$1 ORDER BY started_at ASC NULLS LAST`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var steps []models.WorkflowStep
	for rows.Next() {
		var s models.WorkflowStep
		rows.Scan(&s.ID, &s.JobID, &s.StepName, &s.Status, &s.StartedAt, &s.FinishedAt, &s.OutputKey, &s.ErrorMessage)
		steps = append(steps, s)
	}
	return steps, rows.Err()
}

func (r *PostgresRepository) CreateVideo(ctx context.Context, video *models.Video) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO videos (id,job_id,title,description,duration_seconds,thumbnail_url,youtube_video_id,youtube_embed_url,tigris_video_key,transcript_key,playback_url,status,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		video.ID, video.JobID, video.Title, video.Description, video.DurationSeconds, video.ThumbnailURL,
		video.YouTubeVideoID, video.YouTubeEmbedURL, video.TigrisVideoKey, video.TranscriptKey,
		video.PlaybackURL, video.Status, video.CreatedAt, video.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetVideoByJobID(ctx context.Context, jobID uuid.UUID) (*models.Video, error) {
	v := &models.Video{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id,job_id,title,description,duration_seconds,thumbnail_url,youtube_video_id,youtube_embed_url,tigris_video_key,transcript_key,playback_url,status,created_at,updated_at
		FROM videos WHERE job_id=$1 LIMIT 1`, jobID).Scan(
		&v.ID, &v.JobID, &v.Title, &v.Description, &v.DurationSeconds, &v.ThumbnailURL,
		&v.YouTubeVideoID, &v.YouTubeEmbedURL, &v.TigrisVideoKey, &v.TranscriptKey,
		&v.PlaybackURL, &v.Status, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return v, err
}

func (r *PostgresRepository) UpdateVideo(ctx context.Context, video *models.Video) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE videos SET title=$1,description=$2,duration_seconds=$3,thumbnail_url=$4,
		  youtube_video_id=$5,youtube_embed_url=$6,tigris_video_key=$7,transcript_key=$8,
		  playback_url=$9,status=$10,updated_at=NOW() WHERE id=$11`,
		video.Title, video.Description, video.DurationSeconds, video.ThumbnailURL,
		video.YouTubeVideoID, video.YouTubeEmbedURL, video.TigrisVideoKey, video.TranscriptKey,
		video.PlaybackURL, video.Status, video.ID)
	return err
}

func (r *PostgresRepository) CreateActionCards(ctx context.Context, cards []models.ActionCard) error {
	for _, c := range cards {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO action_cards (id,job_id,title,description,action_type,timestamp_seconds,status,provider,requires_execution,output_key,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			c.ID, c.JobID, c.Title, c.Description, c.ActionType, c.TimestampSeconds,
			c.Status, c.Provider, c.RequiresExecution, c.OutputKey, c.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) ListActionCards(ctx context.Context, jobID uuid.UUID) ([]models.ActionCard, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,job_id,title,description,action_type,timestamp_seconds,status,provider,requires_execution,output_key,created_at
		FROM action_cards WHERE job_id=$1 ORDER BY created_at ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []models.ActionCard
	for rows.Next() {
		var c models.ActionCard
		rows.Scan(&c.ID, &c.JobID, &c.Title, &c.Description, &c.ActionType, &c.TimestampSeconds,
			&c.Status, &c.Provider, &c.RequiresExecution, &c.OutputKey, &c.CreatedAt)
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (r *PostgresRepository) UpdateActionCard(ctx context.Context, card *models.ActionCard) error {
	_, err := r.db.ExecContext(ctx, `UPDATE action_cards SET status=$1,output_key=$2 WHERE id=$3`,
		card.Status, card.OutputKey, card.ID)
	return err
}

func (r *PostgresRepository) CreateClaims(ctx context.Context, claims []models.Claim) error {
	for _, c := range claims {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO claims (id,job_id,claim_text,timestamp_seconds,verification_status,confidence,evidence_key,created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			c.ID, c.JobID, c.ClaimText, c.TimestampSeconds, c.VerificationStatus, c.Confidence, c.EvidenceKey, c.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) ListClaims(ctx context.Context, jobID uuid.UUID) ([]models.Claim, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,job_id,claim_text,timestamp_seconds,verification_status,confidence,evidence_key,created_at
		FROM claims WHERE job_id=$1 ORDER BY created_at ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var claims []models.Claim
	for rows.Next() {
		var c models.Claim
		rows.Scan(&c.ID, &c.JobID, &c.ClaimText, &c.TimestampSeconds, &c.VerificationStatus, &c.Confidence, &c.EvidenceKey, &c.CreatedAt)
		claims = append(claims, c)
	}
	return claims, rows.Err()
}

func (r *PostgresRepository) UpdateClaim(ctx context.Context, claim *models.Claim) error {
	_, err := r.db.ExecContext(ctx, `UPDATE claims SET verification_status=$1,confidence=$2,evidence_key=$3 WHERE id=$4`,
		claim.VerificationStatus, claim.Confidence, claim.EvidenceKey, claim.ID)
	return err
}

func (r *PostgresRepository) CreateExecutionRun(ctx context.Context, run *models.ExecutionRun) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO execution_runs (id,action_card_id,provider,status,input_key,output_key,logs_key,exit_code,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		run.ID, run.ActionCardID, run.Provider, run.Status, run.InputKey, run.OutputKey, run.LogsKey, run.ExitCode, run.CreatedAt)
	return err
}

func (r *PostgresRepository) UpdateExecutionRun(ctx context.Context, run *models.ExecutionRun) error {
	_, err := r.db.ExecContext(ctx, `UPDATE execution_runs SET status=$1,output_key=$2,logs_key=$3,exit_code=$4 WHERE id=$5`,
		run.Status, run.OutputKey, run.LogsKey, run.ExitCode, run.ID)
	return err
}

func (r *PostgresRepository) ListExecutionRuns(ctx context.Context, actionCardID uuid.UUID) ([]models.ExecutionRun, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,action_card_id,provider,status,input_key,output_key,logs_key,exit_code,created_at
		FROM execution_runs WHERE action_card_id=$1 ORDER BY created_at ASC`, actionCardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []models.ExecutionRun
	for rows.Next() {
		var run models.ExecutionRun
		rows.Scan(&run.ID, &run.ActionCardID, &run.Provider, &run.Status, &run.InputKey, &run.OutputKey, &run.LogsKey, &run.ExitCode, &run.CreatedAt)
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (r *PostgresRepository) CreateBrowserRun(ctx context.Context, run *models.BrowserRun) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO browser_runs (id,job_id,action_card_id,provider,status,task,input_key,output_key,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		run.ID, run.JobID, run.ActionCardID, run.Provider, run.Status, run.Task, run.InputKey, run.OutputKey, run.CreatedAt, run.UpdatedAt)
	return err
}

func (r *PostgresRepository) UpdateBrowserRun(ctx context.Context, run *models.BrowserRun) error {
	_, err := r.db.ExecContext(ctx, `UPDATE browser_runs SET status=$1,output_key=$2,updated_at=NOW() WHERE id=$3`,
		run.Status, run.OutputKey, run.ID)
	return err
}

func (r *PostgresRepository) ListBrowserRuns(ctx context.Context, jobID uuid.UUID) ([]models.BrowserRun, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id,job_id,action_card_id,provider,status,task,input_key,output_key,created_at,updated_at
		FROM browser_runs WHERE job_id=$1 ORDER BY created_at ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []models.BrowserRun
	for rows.Next() {
		var run models.BrowserRun
		rows.Scan(&run.ID, &run.JobID, &run.ActionCardID, &run.Provider, &run.Status, &run.Task, &run.InputKey, &run.OutputKey, &run.CreatedAt, &run.UpdatedAt)
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (r *PostgresRepository) ListWorkflowDetails(ctx context.Context, jobID uuid.UUID) (*models.WorkflowDetails, error) {
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
