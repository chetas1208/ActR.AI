package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	Storage  StorageConfig
	DB       DBConfig
	AI       AIConfig
	YouTube  YouTubeConfig
	Daytona  DaytonaConfig
	Rtrvr    RtrvrConfig
	Upload   UploadConfig
	Webhook  WebhookConfig
	Workflow WorkflowConfig
}

type AppConfig struct {
	Name               string
	Env                string
	BaseURL            string
	FrontendBaseURL    string
	CORSAllowedOrigins string
	LogLevel           string
	EnableDebugRoutes  bool
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

type AIConfig struct {
	OpenAIKey              string
	OpenAIBaseURL          string
	Model                  string
	TimeoutSeconds         int
	TranscriptionProvider  string
	TranscriptionModel     string
}

type YouTubeConfig struct {
	APIKey string
}

type DaytonaConfig struct {
	APIKey               string
	APIURL               string
	DefaultImage         string
	CommandTimeout       time.Duration
	DeleteAfterRun       bool
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
	maxUpload, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_BYTES", "524288000"), 10, 64)
	maxDirect, _ := strconv.ParseInt(getEnv("MAX_DIRECT_FILE_BYTES", "524288000"), 10, 64)
	presignUploadTTL, _ := strconv.ParseInt(getEnv("PRESIGNED_UPLOAD_TTL_SECONDS", "900"), 10, 64)
	presignDownloadTTL, _ := strconv.ParseInt(getEnv("PRESIGNED_DOWNLOAD_TTL_SECONDS", "3600"), 10, 64)
	aiTimeout, _ := strconv.Atoi(getEnv("OPENAI_TIMEOUT_SECONDS", "60"))
	daytonaTimeout, _ := strconv.Atoi(getEnv("DAYTONA_COMMAND_TIMEOUT_SECONDS", "60"))
	rtrvrTimeout, _ := strconv.Atoi(getEnv("RTRVR_TIMEOUT_SECONDS", "120"))
	workflowStepTimeout, _ := strconv.Atoi(getEnv("WORKFLOW_STEP_TIMEOUT_SECONDS", "60"))
	workflowMaxRetries, _ := strconv.Atoi(getEnv("WORKFLOW_MAX_RETRIES", "1"))

	return &Config{
		App: AppConfig{
			Name:               getEnv("APP_NAME", "gorube-flow-api"),
			Env:                getEnv("ENVIRONMENT", "development"),
			BaseURL:            getEnv("APP_BASE_URL", "http://localhost:8080"),
			FrontendBaseURL:    getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
			CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
			LogLevel:           getEnv("LOG_LEVEL", "info"),
			EnableDebugRoutes:  getEnvBool("ENABLE_DEBUG_ROUTES", false),
		},
		Storage: StorageConfig{
			Endpoint:           getEnv("TIGRIS_ENDPOINT", "https://t3.storage.dev"),
			AccessKey:          getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey:          getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Region:             getEnv("AWS_REGION", "auto"),
			Bucket:             getEnv("TIGRIS_BUCKET", "gorube-flow"),
			ForcePathStyle:     getEnvBool("S3_FORCE_PATH_STYLE", false),
			PresignUploadTTL:   time.Duration(presignUploadTTL) * time.Second,
			PresignDownloadTTL: time.Duration(presignDownloadTTL) * time.Second,
		},
		DB: DBConfig{
			DatabaseURL:    getEnv("DATABASE_URL", ""),
			InsForgeURL:    getEnv("INSFORGE_API_URL", ""),
			InsForgeAPIKey: getEnv("INSFORGE_API_KEY", ""),
			UseMemory:      getEnvBool("USE_IN_MEMORY_DB", false),
		},
		AI: AIConfig{
			OpenAIKey:             getEnv("OPENAI_API_KEY", ""),
			OpenAIBaseURL:         getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			Model:                 getEnv("OPENAI_MODEL", "gpt-4.1-mini"),
			TimeoutSeconds:        aiTimeout,
			TranscriptionProvider: getEnv("TRANSCRIPTION_PROVIDER", "openai"),
			TranscriptionModel:    getEnv("TRANSCRIPTION_MODEL", "whisper-1"),
		},
		YouTube: YouTubeConfig{
			APIKey: getEnv("YOUTUBE_API_KEY", ""),
		},
		Daytona: DaytonaConfig{
			APIKey:         getEnv("DAYTONA_API_KEY", ""),
			APIURL:         getEnv("DAYTONA_API_URL", "https://app.daytona.io/api"),
			DefaultImage:   getEnv("DAYTONA_DEFAULT_IMAGE", "ubuntu:22.04"),
			CommandTimeout: time.Duration(daytonaTimeout) * time.Second,
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
			StepTimeoutSeconds: workflowStepTimeout,
			MaxRetries:         workflowMaxRetries,
		},
	}
}

func (c *DBConfig) IsInsForge() bool {
	return c.InsForgeURL != "" && c.InsForgeAPIKey != ""
}

func (c *DBConfig) IsPostgres() bool {
	return c.DatabaseURL != ""
}

func (c *DaytonaConfig) IsConfigured() bool {
	return c.APIKey != ""
}

func (c *RtrvrConfig) IsConfigured() bool {
	return c.APIKey != ""
}

func (c *AIConfig) IsConfigured() bool {
	return c.OpenAIKey != ""
}

func (c *YouTubeConfig) IsConfigured() bool {
	return c.APIKey != ""
}

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
