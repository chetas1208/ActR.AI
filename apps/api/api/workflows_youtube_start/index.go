package handler

import (
	"net/http"
	"time"

	"github.com/chetas1208/ActR.AI/apps/api/internal/httpx"
	"github.com/chetas1208/ActR.AI/apps/api/internal/models"
	"github.com/chetas1208/ActR.AI/apps/api/internal/observability"
	"github.com/chetas1208/ActR.AI/apps/api/internal/storage"
	"github.com/chetas1208/ActR.AI/apps/api/internal/youtube"
	"github.com/google/uuid"
)

type youtubeStartRequest struct {
	YoutubeURL string `json:"youtubeUrl"`
}

type youtubeStartResponse struct {
	JobID         string      `json:"jobId"`
	Status        string      `json:"status"`
	RequiresUpload bool       `json:"requiresUpload"`
	Video         interface{} `json:"video,omitempty"`
	Message       string      `json:"message,omitempty"`
}

// Handler is the Vercel function entry point for POST /api/workflows/youtube/start
func Handler(w http.ResponseWriter, r *http.Request) {
	app, err := observability.GetApp()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "service unavailable")
		return
	}
	if httpx.ApplyCORS(w, r, app.Cfg.App.CORSAllowedOrigins) {
		return
	}
	if !httpx.RequireMethod(w, r, http.MethodPost) {
		return
	}

	var req youtubeStartRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.YoutubeURL == "" {
		httpx.WriteValidationError(w, "youtubeUrl", "youtubeUrl is required")
		return
	}
	if !youtube.IsYouTubeURL(req.YoutubeURL) {
		httpx.WriteValidationError(w, "youtubeUrl", "not a valid YouTube URL")
		return
	}

	videoID, err := youtube.ParseVideoID(req.YoutubeURL)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create workflow job
	jobID := uuid.New()
	now := time.Now()
	title := "YouTube Video " + videoID
	job := &models.WorkflowJob{
		ID:           jobID,
		SourceType:   models.SourceTypeYouTube,
		SourceRights: models.SourceRightsYouTubeEmbed,
		SourceURL:    &req.YoutubeURL,
		Title:        &title,
		Status:       models.StatusMetadataFetching,
		Progress:     5,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := app.DB.CreateWorkflowJob(r.Context(), job); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create workflow job")
		return
	}

	// Fetch metadata
	meta, metaErr := app.YouTube.FetchMetadata(r.Context(), videoID)
	if metaErr != nil {
		// Non-fatal: use minimal metadata
		meta = &youtube.Metadata{
			VideoID:  videoID,
			Title:    "YouTube Video " + videoID,
			EmbedURL: youtube.EmbedURL(videoID),
		}
	}

	// Store metadata in Tigris
	keys := storage.ObjectKeys(jobID.String())
	_ = app.Storage.PutJSON(r.Context(), keys["youtube_metadata"], meta)

	// Update video title from metadata
	if meta.Title != "" {
		title = meta.Title
		job.Title = &title
	}

	// Create Video record
	embedURL := meta.EmbedURL
	ytVideoID := meta.VideoID
	videoRecord := &models.Video{
		ID:              uuid.New(),
		JobID:           jobID,
		Title:           meta.Title,
		YouTubeVideoID:  &ytVideoID,
		YouTubeEmbedURL: &embedURL,
		ThumbnailURL:    &meta.ThumbnailURL,
		Status:          "ready",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_ = app.DB.CreateVideo(r.Context(), videoRecord)
	_ = app.DB.UpdateWorkflowJobStatus(r.Context(), jobID, models.StatusSourceReady, "source_ready", 10)

	// YouTube transcript policy: we only use permitted captions/transcript paths.
	// If transcript is unavailable, request user upload.
	// For MVP, always request upload since we don't have OAuth captions access.
	_ = app.DB.UpdateWorkflowJobUserInputRequired(r.Context(), jobID, true, models.RequiredInputTranscriptAudioOrVideo)

	httpx.WriteJSON(w, http.StatusCreated, youtubeStartResponse{
		JobID:          jobID.String(),
		Status:         models.StatusWaitingForUserInput,
		RequiresUpload: true,
		Video: map[string]interface{}{
			"youtubeVideoId":  videoID,
			"youtubeEmbedUrl": embedURL,
			"title":           meta.Title,
			"thumbnailUrl":    meta.ThumbnailURL,
			"channel":         meta.Channel,
		},
		Message: "YouTube metadata fetched. Please upload a transcript, audio, or video file to enable full workflow processing.",
	})
}
