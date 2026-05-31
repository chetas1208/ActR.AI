export type WorkflowStatus =
  | 'created'
  | 'queued'
  | 'metadata_fetching'
  | 'source_ready'
  | 'waiting_for_user_input'
  | 'transcribing'
  | 'chunking'
  | 'summarizing'
  | 'extracting_actions'
  | 'extracting_claims'
  | 'researching_with_rtrvr'
  | 'preparing_execution'
  | 'running_daytona'
  | 'running_rtrvr'
  | 'finalizing'
  | 'ready'
  | 'failed'

export type SourceType = 'youtube' | 'upload' | 'authorized_direct_file'

export interface WorkflowJob {
  id: string
  userId?: string
  sourceType: SourceType
  sourceRights: string
  sourceUrl?: string
  sourceTigrisKey?: string
  transcriptTigrisKey?: string
  title?: string
  status: WorkflowStatus
  currentStep?: string
  progress: number
  requiresUserInput: boolean
  requiredInputType?: string
  errorMessage?: string
  createdAt: string
  updatedAt: string
}

export interface Video {
  id: string
  jobId: string
  title: string
  description?: string
  durationSeconds?: number
  thumbnailUrl?: string
  youtubeVideoId?: string
  youtubeEmbedUrl?: string
  tigrisVideoKey?: string
  transcriptKey?: string
  playbackUrl?: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface WorkflowStep {
  id: string
  jobId: string
  stepName: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  startedAt?: string
  finishedAt?: string
  outputKey?: string
  errorMessage?: string
}

export interface ActionCard {
  id: string
  jobId: string
  title: string
  description: string
  actionType: 'checklist' | 'code' | 'research' | 'study_plan' | 'browser_action'
  timestampSeconds?: number
  status: 'pending' | 'running' | 'completed' | 'failed'
  provider?: string
  requiresExecution: boolean
  outputKey?: string
  createdAt: string
}

export interface Claim {
  id: string
  jobId: string
  claimText: string
  timestampSeconds?: number
  verificationStatus: 'pending' | 'researched' | 'uncertain' | 'failed'
  confidence?: number
  evidenceKey?: string
  createdAt: string
}

export interface ExecutionRun {
  id: string
  actionCardId: string
  provider: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  inputKey?: string
  outputKey?: string
  logsKey?: string
  exitCode?: number
  createdAt: string
}

export interface BrowserRun {
  id: string
  jobId: string
  actionCardId?: string
  provider: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  task: string
  inputKey?: string
  outputKey?: string
  createdAt: string
  updatedAt: string
}

export interface WorkflowDetails {
  job: WorkflowJob
  video?: Video
  steps: WorkflowStep[]
  actionCards: ActionCard[]
  claims: Claim[]
  executions: ExecutionRun[]
  browserRuns: BrowserRun[]
  artifactUrls?: Record<string, string>
}

export interface PresignUploadResponse {
  jobId: string
  uploadUrl: string
  objectKey: string
}

export interface YouTubeStartResponse {
  jobId: string
  status: WorkflowStatus
  requiresUpload: boolean
  video?: {
    youtubeVideoId: string
    youtubeEmbedUrl: string
    title: string
    thumbnailUrl: string
    channel: string
  }
  message?: string
}

export interface ApiError {
  error: string
  field?: string
  detail?: string
  provider?: string
}

export const TERMINAL_STATUSES: WorkflowStatus[] = ['ready', 'failed', 'waiting_for_user_input']

export const STATUS_LABELS: Record<WorkflowStatus, string> = {
  created: 'Created',
  queued: 'Queued',
  metadata_fetching: 'Fetching Metadata',
  source_ready: 'Source Ready',
  waiting_for_user_input: 'Waiting for Input',
  transcribing: 'Transcribing',
  chunking: 'Chunking',
  summarizing: 'Summarizing',
  extracting_actions: 'Extracting Actions',
  extracting_claims: 'Extracting Claims',
  researching_with_rtrvr: 'Browser Research',
  preparing_execution: 'Preparing Execution',
  running_daytona: 'Running Code',
  running_rtrvr: 'Running Browser Task',
  finalizing: 'Finalizing',
  ready: 'Ready',
  failed: 'Failed',
}

export const STATUS_PROGRESS: Record<WorkflowStatus, number> = {
  created: 0,
  queued: 5,
  metadata_fetching: 10,
  source_ready: 15,
  waiting_for_user_input: 15,
  transcribing: 25,
  chunking: 35,
  summarizing: 50,
  extracting_actions: 65,
  extracting_claims: 75,
  researching_with_rtrvr: 85,
  preparing_execution: 88,
  running_daytona: 92,
  running_rtrvr: 92,
  finalizing: 97,
  ready: 100,
  failed: 0,
}
