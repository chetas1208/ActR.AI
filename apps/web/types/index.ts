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
export type SourceMode  = 'youtube' | 'upload' | 'direct'

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

export interface ProviderStatus {
  nvidiaLLM: boolean
  nvidiaASR: boolean
  tigris: boolean
  daytona: boolean
  rtrvr: boolean
  youtube: boolean
  database: 'insforge' | 'postgres' | 'memory' | 'missing'
}

export const TERMINAL_STATUSES: WorkflowStatus[] = ['ready', 'failed', 'waiting_for_user_input']

export const STATUS_LABELS: Record<WorkflowStatus, string> = {
  created:               'Created',
  queued:                'Queued',
  metadata_fetching:     'Fetching Metadata',
  source_ready:          'Source Ready',
  waiting_for_user_input: 'Waiting for Input',
  transcribing:          'Transcribing',
  chunking:              'Processing',
  summarizing:           'Summarizing',
  extracting_actions:    'Extracting Actions',
  extracting_claims:     'Extracting Claims',
  researching_with_rtrvr: 'Browser Research',
  preparing_execution:   'Preparing Execution',
  running_daytona:       'Running Code',
  running_rtrvr:         'Browser Automation',
  finalizing:            'Finalizing',
  ready:                 'Ready',
  failed:                'Failed',
}

export const STATUS_COLOR: Record<WorkflowStatus, string> = {
  created:               'text-zinc-400',
  queued:                'text-zinc-400',
  metadata_fetching:     'text-electric-400',
  source_ready:          'text-electric-400',
  waiting_for_user_input: 'text-violet-400',
  transcribing:          'text-cyan-400',
  chunking:              'text-cyan-400',
  summarizing:           'text-cyan-400',
  extracting_actions:    'text-cyan-400',
  extracting_claims:     'text-cyan-400',
  researching_with_rtrvr: 'text-violet-400',
  preparing_execution:   'text-electric-400',
  running_daytona:       'text-electric-400',
  running_rtrvr:         'text-violet-400',
  finalizing:            'text-cyan-400',
  ready:                 'text-emerald-400',
  failed:                'text-red-400',
}
