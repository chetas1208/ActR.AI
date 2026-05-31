package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Metadata holds extracted YouTube video metadata.
type Metadata struct {
	VideoID     string `json:"videoId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Channel     string `json:"channel"`
	PublishedAt string `json:"publishedAt,omitempty"`
	Duration    string `json:"duration,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	EmbedURL    string `json:"embedUrl"`
	OEmbedTitle string `json:"oEmbedTitle,omitempty"`
}

var youtubeIDRegex = regexp.MustCompile(
	`(?:youtube\.com\/(?:watch\?v=|shorts\/|embed\/|v\/)|youtu\.be\/)([A-Za-z0-9_-]{11})`,
)

// ParseVideoID extracts the YouTube video ID from a URL.
func ParseVideoID(rawURL string) (string, error) {
	m := youtubeIDRegex.FindStringSubmatch(rawURL)
	if len(m) < 2 {
		return "", fmt.Errorf("could not extract YouTube video ID from URL: %s", rawURL)
	}
	return m[1], nil
}

// EmbedURL returns the standard YouTube embed URL for a video ID.
func EmbedURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
}

// Client fetches YouTube metadata using the Data API v3 or oEmbed fallback.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchMetadata fetches video metadata. Uses Data API v3 if apiKey is set, oEmbed otherwise.
func (c *Client) FetchMetadata(ctx context.Context, videoID string) (*Metadata, error) {
	if c.apiKey != "" {
		return c.fetchFromDataAPI(ctx, videoID)
	}
	return c.fetchFromOEmbed(ctx, videoID)
}

func (c *Client) fetchFromDataAPI(ctx context.Context, videoID string) (*Metadata, error) {
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?id=%s&part=snippet,contentDetails&key=%s",
		url.QueryEscape(videoID), url.QueryEscape(c.apiKey),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube data api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube data api returned status %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				ChannelTitle string `json:"channelTitle"`
				PublishedAt string `json:"publishedAt"`
				Thumbnails  struct {
					High struct {
						URL string `json:"url"`
					} `json:"high"`
					MaxRes struct {
						URL string `json:"url"`
					} `json:"maxres"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode youtube response: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("youtube video not found: %s", videoID)
	}

	item := result.Items[0]
	thumbnail := item.Snippet.Thumbnails.MaxRes.URL
	if thumbnail == "" {
		thumbnail = item.Snippet.Thumbnails.High.URL
	}

	return &Metadata{
		VideoID:      videoID,
		Title:        item.Snippet.Title,
		Description:  item.Snippet.Description,
		Channel:      item.Snippet.ChannelTitle,
		PublishedAt:  item.Snippet.PublishedAt,
		Duration:     item.ContentDetails.Duration,
		ThumbnailURL: thumbnail,
		EmbedURL:     EmbedURL(videoID),
	}, nil
}

// oEmbed response shape
type oEmbedResponse struct {
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

func (c *Client) fetchFromOEmbed(ctx context.Context, videoID string) (*Metadata, error) {
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	apiURL := fmt.Sprintf("https://www.youtube.com/oembed?url=%s&format=json", url.QueryEscape(videoURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oembed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &Metadata{
			VideoID:  videoID,
			Title:    "YouTube Video " + videoID,
			EmbedURL: EmbedURL(videoID),
		}, nil
	}

	var oembed oEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&oembed); err != nil {
		return nil, fmt.Errorf("decode oembed: %w", err)
	}

	return &Metadata{
		VideoID:      videoID,
		Title:        oembed.Title,
		Channel:      oembed.AuthorName,
		ThumbnailURL: oembed.ThumbnailURL,
		EmbedURL:     EmbedURL(videoID),
		OEmbedTitle:  oembed.Title,
	}, nil
}

// IsConfigured returns true if a YouTube Data API key is set.
func (c *Client) IsConfigured() bool {
	return c.apiKey != ""
}

// IsYouTubeURL returns true if the URL is a recognizable YouTube URL.
func IsYouTubeURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return strings.HasSuffix(host, "youtube.com") || host == "youtu.be"
}
