package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App          AppConfig
	Storage      StorageConfig
	DB           DBConfig
	NVIDIA       NVIDIAConfig
	NVIDIASR     NVIDIASRConfig
	Transcription TranscriptionConfig
	YouTube      YouTubeConfig
	Daytona      DaytonaConfig
	Rtrvr        RtrvrConfig
	Upload       UploadConfig
	Webhook      WebhookConfig
	Workflow     WorkflowConfig
}

type AppConfig struct {
	Name               string
	Env                string
	BaseURL            string
	FrontendBaseURL    string
	CORSAllowedOrigins string
	LogLevel           string
	EnableDebugRoutes  bool
	EnableProviderStatus bool
}

type StorageConfig struct {
	Endpoint           string
	AccessKey          string
	SecretKey          string
	Region             string
	Bucket             string
	ForcePathStyle     bool
	PresignUploadTTL   time.Duration
	PresignDownloadTTL time.Duration
}

type DBConfig struct {
	DatabaseURL    string
	InsForgeURL    string
	InsForgeAPIKey string
	UseMemory      bool
}

// NVIDIAConfig holds configuration for NVIDIA NIM LLM calls.
type NVIDIAConfig struct {
	APIKey        string
	BaseURL       string
	LLMModel      string
	FastModel     string
	CodingModel   string
	TimeoutSeconds int
}

// NVIDIASRConfig holds configuration for NVIDIA ASR transcription.
type NVIDIASRConfig struct {
	Provider       string // "nvidia"
	APIKey         string
	BaseURL        string
	Model          string
	Language       string
	TimeoutSeconds int
}

type TranscriptionConfig struct {
	Provider               string // "nvidia" | "local_whisper"
	LocalEnabled           bool
	LocalWhisperModelPath  string
	LocalWhisperServerURL  string
}

type YouTubeConfig struct {
	APIKey string
}

type DaytonaConfig struct {
	APIKey         string
	APIURL         string
	DefaultImage   string
	CommandTimeout time.Duration
	DeleteAfterRun bool
}

type RtrvrConfig struct {
	APIKey         string
	APIURL         string
	Mode           string
	TimeoutSeconds int
}

type UploadConfig struct {
	MaxUploadBytes     int64
	MaxDirectFileBytes int64
	AllowedVideoTypes  []string
}

type WebhookConfig struct {
	Secret string
}

type WorkflowConfig struct {
	AutoAdvance        bool
	StepTimeoutSeconds int
	MaxRetries         int
}

func Load() *Config {
	maxUpload, _  := strconv.ParseInt(getEnv("MAX_UPLOAD_BYTES", "524288000"), 10, 64)
	maxDirect, _  := strconv.ParseInt(getEnv("MAX_DIRECT_FILE_BYTES", "524288000"), 10, 64)
	presignUp, _  := strconv.ParseInt(getEnv("PRESIGNED_UPLOAD_TTL_SECONDS", "900"), 10, 64)
	presignDl, _  := strconv.ParseInt(getEnv("PRESIGNED_DOWNLOAD_TTL_SECONDS", "3600"), 10, 64)
	nimTimeout, _ := strconv.Atoi(getEnv("NIM_TIMEOUT_SECONDS", "60"))
	asrTimeout, _ := strconv.Atoi(getEnv("NVIDIA_ASR_TIMEOUT_SECONDS", "120"))
	dayTimeout, _ := strconv.Atoi(getEnv("DAYTONA_COMMAND_TIMEOUT_SECONDS", "60"))
	rtrvrTimeout, _ := strconv.Atoi(getEnv("RTRVR_TIMEOUT_SECONDS", "120"))
	workflowStep, _ := strconv.Atoi(getEnv("WORKFLOW_STEP_TIMEOUT_SECONDS", "60"))
	workflowRetry, _ := strconv.Atoi(getEnv("WORKFLOW_MAX_RETRIES", "1"))

	nimAPIKey  := getEnv("NIM_API_KEY", "")
	asrAPIKey  := getEnv("NVIDIA_ASR_API_KEY", nimAPIKey) // default to NIM key

	nimLLM     := getEnv("NIM_LLM_MODEL", "nvidia/nemotron-3-super-120b-a12b")
	nimFast    := getEnv("NIM_FAST_MODEL", "nvidia/llama-3.1-nemotron-nano-8b-v1")
	nimCoding  := getEnv("NIM_CODING_MODEL", nimLLM) // defaults to LLM model

	return &Config{
		App: AppConfig{
			Name:                 getEnv("APP_NAME", "actr-ai-api"),
			Env:                  getEnv("ENVIRONMENT", "development"),
			BaseURL:              getEnv("APP_BASE_URL", "http://localhost:8080"),
			FrontendBaseURL:      getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
			CORSAllowedOrigins:   getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
			LogLevel:             getEnv("LOG_LEVEL", "info"),
			EnableDebugRoutes:    getEnvBool("ENABLE_DEBUG_ROUTES", false),
			EnableProviderStatus: getEnvBool("ENABLE_PROVIDER_STATUS_ENDPOINT", true),
		},
		Storage: StorageConfig{
			Endpoint:           getEnv("TIGRIS_ENDPOINT", "https://t3.storage.dev"),
			AccessKey:          getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey:          getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Region:             getEnv("AWS_REGION", "auto"),
			Bucket:             getEnv("TIGRIS_BUCKET", "actr-ai"),
			ForcePathStyle:     getEnvBool("S3_FORCE_PATH_STYLE", false),
			PresignUploadTTL:   time.Duration(presignUp) * time.Second,
			PresignDownloadTTL: time.Duration(presignDl) * time.Second,
		},
		DB: DBConfig{
			DatabaseURL:    getEnv("DATABASE_URL", ""),
			InsForgeURL:    getEnv("INSFORGE_API_URL", ""),
			InsForgeAPIKey: getEnv("INSFORGE_API_KEY", ""),
			UseMemory:      getEnvBool("USE_IN_MEMORY_DB", false),
		},
		NVIDIA: NVIDIAConfig{
			APIKey:         nimAPIKey,
			BaseURL:        getEnv("NIM_BASE_URL", "https://integrate.api.nvidia.com/v1"),
			LLMModel:       nimLLM,
			FastModel:      nimFast,
			CodingModel:    nimCoding,
			TimeoutSeconds: nimTimeout,
		},
		NVIDIASR: NVIDIASRConfig{
			Provider:       getEnv("NVIDIA_ASR_PROVIDER", "nvidia"),
			APIKey:         asrAPIKey,
			BaseURL:        getEnv("NVIDIA_ASR_BASE_URL", "https://integrate.api.nvidia.com/v1"),
			Model:          getEnv("NVIDIA_ASR_MODEL", "nvidia/parakeet-1.1b-rnnt-multilingual-asr"),
			Language:       getEnv("NVIDIA_ASR_LANGUAGE", "en"),
			TimeoutSeconds: asrTimeout,
		},
		Transcription: TranscriptionConfig{
			Provider:              getEnv("TRANSCRIPTION_PROVIDER", "nvidia"),
			LocalEnabled:          getEnvBool("LOCAL_TRANSCRIPTION_ENABLED", false),
			LocalWhisperModelPath: getEnv("LOCAL_WHISPER_MODEL_PATH", ""),
			LocalWhisperServerURL: getEnv("LOCAL_WHISPER_SERVER_URL", ""),
		},
		YouTube: YouTubeConfig{
			APIKey: getEnv("YOUTUBE_API_KEY", ""),
		},
		Daytona: DaytonaConfig{
			APIKey:         getEnv("DAYTONA_API_KEY", ""),
			APIURL:         getEnv("DAYTONA_API_URL", "https://app.daytona.io/api"),
			DefaultImage:   getEnv("DAYTONA_DEFAULT_IMAGE", "ubuntu:22.04"),
			CommandTimeout: time.Duration(dayTimeout) * time.Second,
			DeleteAfterRun: getEnvBool("DAYTONA_DELETE_SANDBOX_AFTER_RUN", true),
		},
		Rtrvr: RtrvrConfig{
			APIKey:         getEnv("RTRVR_API_KEY", ""),
			APIURL:         getEnv("RTRVR_API_URL", "https://api.rtrvr.ai"),
			Mode:           getEnv("RTRVR_MODE", "cloud"),
			TimeoutSeconds: rtrvrTimeout,
		},
		Upload: UploadConfig{
			MaxUploadBytes:     maxUpload,
			MaxDirectFileBytes: maxDirect,
			AllowedVideoTypes: strings.Split(
				getEnv("ALLOWED_VIDEO_TYPES", "video/mp4,video/quicktime,video/webm,audio/mpeg,audio/mp4,audio/wav,text/plain,text/vtt,application/json"),
				",",
			),
		},
		Webhook: WebhookConfig{
			Secret: getEnv("WEBHOOK_SECRET", ""),
		},
		Workflow: WorkflowConfig{
			AutoAdvance:        getEnvBool("WORKFLOW_AUTO_ADVANCE", true),
			StepTimeoutSeconds: workflowStep,
			MaxRetries:         workflowRetry,
		},
	}
}

// Provider availability helpers
func (c *NVIDIAConfig) IsConfigured() bool { return c.APIKey != "" }
func (c *NVIDIASRConfig) IsConfigured() bool { return c.APIKey != "" }
func (c *YouTubeConfig) IsConfigured() bool  { return c.APIKey != "" }
func (c *DaytonaConfig) IsConfigured() bool  { return c.APIKey != "" }
func (c *RtrvrConfig) IsConfigured() bool    { return c.APIKey != "" }
func (c *DBConfig) IsInsForge() bool         { return c.InsForgeURL != "" && c.InsForgeAPIKey != "" }
func (c *DBConfig) IsPostgres() bool         { return c.DatabaseURL != "" }
func (c *StorageConfig) IsConfigured() bool  { return c.AccessKey != "" && c.SecretKey != "" && c.Bucket != "" }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
