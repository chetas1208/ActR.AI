package models

import (
	"time"

	"github.com/google/uuid"
)

// Source types
const (
	SourceTypeYouTube          = "youtube"
	SourceTypeUpload           = "upload"
	SourceTypeAuthorizedDirect = "authorized_direct_file"

	SourceRightsYouTubeEmbed    = "youtube_embed_only"
	SourceRightsUploadedByUser  = "uploaded_by_user"
	SourceRightsAuthorizedDirect = "authorized_direct_file"
)

// Workflow job statuses
const (
	StatusCreated              = "created"
	StatusQueued               = "queued"
	StatusMetadataFetching     = "metadata_fetching"
	StatusSourceReady          = "source_ready"
	StatusWaitingForUserInput  = "waiting_for_user_input"
	StatusTranscribing         = "transcribing"
	StatusChunking             = "chunking"
	StatusSummarizing          = "summarizing"
	StatusExtractingActions    = "extracting_actions"
	StatusExtractingClaims     = "extracting_claims"
	StatusResearchingWithRtrvr = "researching_with_rtrvr"
	StatusPreparingExecution   = "preparing_execution"
	StatusRunningDaytona       = "running_daytona"
	StatusRunningRtrvr         = "running_rtrvr"
	StatusFinalizing           = "finalizing"
	StatusReady                = "ready"
	StatusFailed               = "failed"
)

// Required input types
const (
	RequiredInputTranscriptAudioOrVideo = "transcript_audio_or_video"
	RequiredInputTranscript             = "transcript"
)

// Action card types
const (
	ActionTypeChecklist     = "checklist"
	ActionTypeCode          = "code"
	ActionTypeResearch      = "research"
	ActionTypeStudyPlan     = "study_plan"
	ActionTypeBrowserAction = "browser_action"
)

// Action/step statuses
const (
	ActionStatusPending   = "pending"
	ActionStatusRunning   = "running"
	ActionStatusCompleted = "completed"
	ActionStatusFailed    = "failed"
)

// Claim verification statuses
const (
	ClaimStatusPending    = "pending"
	ClaimStatusResearched = "researched"
	ClaimStatusUncertain  = "uncertain"
	ClaimStatusFailed     = "failed"
)

// Execution/browser run statuses
const (
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusFailed    = "failed"
)

// WorkflowJob is the central entity for an ActR.AI workflow job.
type WorkflowJob struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	UserID              *uuid.UUID `json:"userId,omitempty" db:"user_id"`
	SourceType          string     `json:"sourceType" db:"source_type"`
	SourceRights        string     `json:"sourceRights" db:"source_rights"`
	SourceURL           *string    `json:"sourceUrl,omitempty" db:"source_url"`
	SourceTigrisKey     *string    `json:"sourceTigrisKey,omitempty" db:"source_tigris_key"`
	TranscriptTigrisKey *string    `json:"transcriptTigrisKey,omitempty" db:"transcript_tigris_key"`
	Title               *string    `json:"title,omitempty" db:"title"`
	Status              string     `json:"status" db:"status"`
	CurrentStep         *string    `json:"currentStep,omitempty" db:"current_step"`
	Progress            int        `json:"progress" db:"progress"`
	RequiresUserInput   bool       `json:"requiresUserInput" db:"requires_user_input"`
	RequiredInputType   *string    `json:"requiredInputType,omitempty" db:"required_input_type"`
	ErrorMessage        *string    `json:"errorMessage,omitempty" db:"error_message"`
	CreatedAt           time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time  `json:"updatedAt" db:"updated_at"`
}

// Video holds metadata about a video associated with a workflow job.
type Video struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	JobID           uuid.UUID  `json:"jobId" db:"job_id"`
	Title           string     `json:"title" db:"title"`
	Description     *string    `json:"description,omitempty" db:"description"`
	DurationSeconds *int       `json:"durationSeconds,omitempty" db:"duration_seconds"`
	ThumbnailURL    *string    `json:"thumbnailUrl,omitempty" db:"thumbnail_url"`
	YouTubeVideoID  *string    `json:"youtubeVideoId,omitempty" db:"youtube_video_id"`
	YouTubeEmbedURL *string    `json:"youtubeEmbedUrl,omitempty" db:"youtube_embed_url"`
	TigrisVideoKey  *string    `json:"tigrisVideoKey,omitempty" db:"tigris_video_key"`
	TranscriptKey   *string    `json:"transcriptKey,omitempty" db:"transcript_key"`
	PlaybackURL     *string    `json:"playbackUrl,omitempty" db:"playback_url"`
	Status          string     `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time  `json:"updatedAt" db:"updated_at"`
}

// WorkflowStep tracks progress of individual pipeline steps.
type WorkflowStep struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	JobID        uuid.UUID  `json:"jobId" db:"job_id"`
	StepName     string     `json:"stepName" db:"step_name"`
	Status       string     `json:"status" db:"status"`
	StartedAt    *time.Time `json:"startedAt,omitempty" db:"started_at"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty" db:"finished_at"`
	OutputKey    *string    `json:"outputKey,omitempty" db:"output_key"`
	ErrorMessage *string    `json:"errorMessage,omitempty" db:"error_message"`
}

// ActionCard represents an actionable item extracted from a workflow.
type ActionCard struct {
	ID                uuid.UUID `json:"id" db:"id"`
	JobID             uuid.UUID `json:"jobId" db:"job_id"`
	Title             string    `json:"title" db:"title"`
	Description       string    `json:"description" db:"description"`
	ActionType        string    `json:"actionType" db:"action_type"`
	TimestampSeconds  *int      `json:"timestampSeconds,omitempty" db:"timestamp_seconds"`
	Status            string    `json:"status" db:"status"`
	Provider          *string   `json:"provider,omitempty" db:"provider"`
	RequiresExecution bool      `json:"requiresExecution" db:"requires_execution"`
	OutputKey         *string   `json:"outputKey,omitempty" db:"output_key"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
}

// Claim is a factual assertion extracted from the transcript/summary.
type Claim struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	JobID              uuid.UUID  `json:"jobId" db:"job_id"`
	ClaimText          string     `json:"claimText" db:"claim_text"`
	TimestampSeconds   *int       `json:"timestampSeconds,omitempty" db:"timestamp_seconds"`
	VerificationStatus string     `json:"verificationStatus" db:"verification_status"`
	Confidence         *float64   `json:"confidence,omitempty" db:"confidence"`
	EvidenceKey        *string    `json:"evidenceKey,omitempty" db:"evidence_key"`
	CreatedAt          time.Time  `json:"createdAt" db:"created_at"`
}

// ExecutionRun tracks a Daytona code execution.
type ExecutionRun struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ActionCardID uuid.UUID  `json:"actionCardId" db:"action_card_id"`
	Provider     string     `json:"provider" db:"provider"`
	Status       string     `json:"status" db:"status"`
	InputKey     *string    `json:"inputKey,omitempty" db:"input_key"`
	OutputKey    *string    `json:"outputKey,omitempty" db:"output_key"`
	LogsKey      *string    `json:"logsKey,omitempty" db:"logs_key"`
	ExitCode     *int       `json:"exitCode,omitempty" db:"exit_code"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

// BrowserRun tracks a Rtrvr.ai browser automation/research task.
type BrowserRun struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	JobID        uuid.UUID  `json:"jobId" db:"job_id"`
	ActionCardID *uuid.UUID `json:"actionCardId,omitempty" db:"action_card_id"`
	Provider     string     `json:"provider" db:"provider"`
	Status       string     `json:"status" db:"status"`
	Task         string     `json:"task" db:"task"`
	InputKey     *string    `json:"inputKey,omitempty" db:"input_key"`
	OutputKey    *string    `json:"outputKey,omitempty" db:"output_key"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

// WorkflowDetails aggregates all data for a workflow job.
type WorkflowDetails struct {
	Job          *WorkflowJob      `json:"job"`
	Video        *Video            `json:"video,omitempty"`
	Steps        []WorkflowStep    `json:"steps"`
	ActionCards  []ActionCard      `json:"actionCards"`
	Claims       []Claim           `json:"claims"`
	Executions   []ExecutionRun    `json:"executions"`
	BrowserRuns  []BrowserRun      `json:"browserRuns"`
	ArtifactURLs map[string]string `json:"artifactUrls,omitempty"`
}

// AI-generated artifact types

type SummaryJSON struct {
	Title      string        `json:"title"`
	Summary    string        `json:"summary"`
	Audience   string        `json:"audience"`
	Difficulty string        `json:"difficulty"`
	KeyTopics  []string      `json:"keyTopics"`
	Chapters   []ChapterItem `json:"chapters"`
}

type ChapterItem struct {
	Title        string `json:"title"`
	StartSeconds int    `json:"startSeconds"`
	EndSeconds   int    `json:"endSeconds"`
	Summary      string `json:"summary"`
}

type ActionCardsJSON struct {
	Actions []ActionCardItem `json:"actions"`
}

type ActionCardItem struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	Type              string `json:"type"`
	TimestampSeconds  int    `json:"timestampSeconds"`
	RequiresExecution bool   `json:"requiresExecution"`
	Provider          string `json:"provider"`
}

type ClaimsJSON struct {
	Claims []ClaimItem `json:"claims"`
}

type ClaimItem struct {
	Text              string `json:"text"`
	TimestampSeconds  int    `json:"timestampSeconds"`
	NeedsVerification bool   `json:"needsVerification"`
	SearchQuery       string `json:"searchQuery"`
}

// RtrvrResearchTaskJSON is the structured task spec sent to Rtrvr.
type RtrvrResearchTaskJSON struct {
	Task       string              `json:"task"`
	TargetURLs []string            `json:"targetUrls"`
	Questions  []string            `json:"questions"`
	Expected   RtrvrExpectedOutput `json:"expectedOutput"`
}

type RtrvrExpectedOutput struct {
	Sources          []string `json:"sources"`
	ClaimAssessments []string `json:"claimAssessments"`
	Notes            []string `json:"notes"`
}
