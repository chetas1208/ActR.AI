package validation

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

var allowedExtensions = map[string]bool{
	".mp4": true, ".mov": true, ".webm": true, ".m4a": true, ".mp3": true,
	".wav": true, ".vtt": true, ".srt": true, ".txt": true, ".json": true,
}

var allowedContentTypes = map[string]bool{
	"video/mp4": true, "video/quicktime": true, "video/webm": true,
	"audio/mpeg": true, "audio/mp4": true, "audio/wav": true,
	"audio/ogg": true, "text/plain": true, "text/vtt": true,
	"application/json": true, "application/x-subrip": true,
}

// ValidateDirectFileURL validates a user-provided direct download URL for safety.
// It rejects private networks, insecure schemes, and disallowed file types.
// It follows redirects carefully and revalidates the final URL.
func ValidateDirectFileURL(ctx context.Context, rawURL string, maxBytes int64) (*URLInfo, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if strings.ToLower(parsed.Scheme) != "https" {
		return nil, fmt.Errorf("only HTTPS URLs are allowed")
	}

	host := strings.ToLower(parsed.Hostname())
	if isPrivateHost(host) {
		return nil, fmt.Errorf("private, localhost, and internal URLs are not allowed")
	}

	ext := strings.ToLower(filepath.Ext(parsed.Path))
	if ext != "" && !allowedExtensions[ext] {
		return nil, fmt.Errorf("file extension %q is not allowed; permitted: mp4, mov, webm, m4a, mp3, wav, vtt, srt, txt, json", ext)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 5 {
				return fmt.Errorf("too many redirects")
			}
			finalHost := strings.ToLower(req.URL.Hostname())
			if isPrivateHost(finalHost) {
				return fmt.Errorf("redirect to private host %q is not allowed", finalHost)
			}
			if strings.ToLower(req.URL.Scheme) != "https" {
				return fmt.Errorf("redirect to non-HTTPS URL is not allowed")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build HEAD request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HEAD request failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("URL returned HTTP %d", resp.StatusCode)
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	ct = strings.SplitN(ct, ";", 2)[0]
	ct = strings.TrimSpace(ct)

	if ct != "" && !allowedContentTypes[ct] {
		return nil, fmt.Errorf("content type %q is not allowed", ct)
	}

	cl := resp.ContentLength
	if cl > 0 && maxBytes > 0 && cl > maxBytes {
		return nil, fmt.Errorf("file size %d bytes exceeds maximum allowed %d bytes", cl, maxBytes)
	}

	return &URLInfo{
		FinalURL:    resp.Request.URL.String(),
		ContentType: ct,
		SizeBytes:   cl,
	}, nil
}

// URLInfo holds validated information about a direct file URL.
type URLInfo struct {
	FinalURL    string
	ContentType string
	SizeBytes   int64
}

// InferContentType guesses the MIME type from a file extension.
func InferContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	case ".wav":
		return "audio/wav"
	case ".vtt":
		return "text/vtt"
	case ".srt":
		return "application/x-subrip"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

// IsTranscriptContentType returns true if the content type represents a transcript.
func IsTranscriptContentType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]))
	switch ct {
	case "text/vtt", "application/x-subrip", "text/plain", "application/json":
		return true
	}
	return false
}

// IsTranscriptExt returns true if the extension is a transcript file extension.
func IsTranscriptExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".vtt", ".srt", ".txt", ".json":
		return true
	}
	return false
}

var privateRanges = []string{
	"10.", "172.16.", "172.17.", "172.18.", "172.19.", "172.20.", "172.21.",
	"172.22.", "172.23.", "172.24.", "172.25.", "172.26.", "172.27.", "172.28.",
	"172.29.", "172.30.", "172.31.", "192.168.", "127.", "0.", "169.254.",
}

func isPrivateHost(host string) bool {
	if host == "localhost" || host == "::1" || host == "" {
		return true
	}
	// Strip port
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	for _, prefix := range privateRanges {
		if strings.HasPrefix(host, prefix) {
			return true
		}
	}
	// Check for internal TLDs
	if strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".corp") {
		return true
	}
	return false
}
