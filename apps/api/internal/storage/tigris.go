package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	gcfg "github.com/chetas1208/gorube-flow/api/internal/config"
)

// Client wraps S3-compatible Tigris operations.
type Client struct {
	s3     *s3.Client
	pre    *s3.PresignClient
	bucket string
}

// New creates a Tigris-backed storage client from the app config.
func New(cfg *gcfg.Config) (*Client, error) {
	endpoint := strings.TrimRight(cfg.Storage.Endpoint, "/")
	if endpoint == "" {
		endpoint = "https://fly.storage.tigris.dev"
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
	})

	region := cfg.Storage.Region
	if region == "" {
		region = "auto"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithEndpointResolverWithOptions(resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.Storage.AccessKey, cfg.Storage.SecretKey, "",
		)),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("build storage config: %w", err)
	}

	s3c := s3.NewFromConfig(awsCfg)
	return &Client{
		s3:     s3c,
		pre:    s3.NewPresignClient(s3c),
		bucket: cfg.Storage.Bucket,
	}, nil
}

// GeneratePresignedUploadURL creates a presigned PUT URL for direct browser upload.
func (c *Client) GeneratePresignedUploadURL(ctx context.Context, key, contentType string) (string, error) {
	req, err := c.pre.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(1*time.Hour))
	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}
	return req.URL, nil
}

// SignedObjectURL creates a presigned GET URL for reading a private object.
func (c *Client) SignedObjectURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := c.pre.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}
	return req.URL, nil
}

// ObjectURL returns the canonical object URL (not signed) for public-read objects.
func (c *Client) ObjectURL(key string) string {
	// Tigris public URL pattern
	return fmt.Sprintf("https://%s.fly.storage.tigris.dev/%s", c.bucket, key)
}

// PutJSON serialises v as JSON and stores it at key.
func (c *Client) PutJSON(ctx context.Context, key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	return err
}

// GetJSON fetches key and deserialises JSON into dst.
func (c *Client) GetJSON(ctx context.Context, key string, dst interface{}) error {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("get object %s: %w", key, err)
	}
	defer out.Body.Close()
	return json.NewDecoder(out.Body).Decode(dst)
}

// StreamUpload streams r directly into Tigris at key.
func (c *Client) StreamUpload(ctx context.Context, key string, r io.Reader, contentType string) error {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	return err
}

// StoreArtifact stores a JSON-serialisable artifact at a structured key path.
func (c *Client) StoreArtifact(ctx context.Context, jobID, category, filename string, v interface{}) (string, error) {
	key := fmt.Sprintf("videos/%s/%s/%s", jobID, category, filename)
	if err := c.PutJSON(ctx, key, v); err != nil {
		return "", err
	}
	return key, nil
}

// GetObjectBytes downloads an object and returns its raw bytes.
func (c *Client) GetObjectBytes(ctx context.Context, key string) ([]byte, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %s: %w", key, err)
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}

// HeadObject checks whether an object exists.
func (c *Client) HeadObject(ctx context.Context, key string) error {
	_, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

// ObjectKeys returns canonical Tigris key paths for a workflow job.
func ObjectKeys(jobID string) map[string]string {
	base := "videos/" + jobID
	return map[string]string{
		"source_video":      base + "/source/original.mp4",
		"source_audio":      base + "/source/audio.mp3",
		"source_transcript": base + "/source/transcript.txt",
		"source_srt":        base + "/source/transcript.srt",
		"source_vtt":        base + "/source/transcript.vtt",
		"youtube_metadata":  base + "/youtube/metadata.json",
		"youtube_embed":     base + "/youtube/embed.json",
		"transcript_json":   base + "/transcript/transcript.json",
		"transcript_vtt":    base + "/transcript/transcript.vtt",
		"summary":           base + "/analysis/summary.json",
		"action_cards":      base + "/analysis/action_cards.json",
		"claims":            base + "/analysis/claims.json",
		"rtrvr_source_research": base + "/rtrvr/source_research.json",
		"daytona_logs":      base + "/daytona/execution_logs.json",
		"rtrvr_results":     base + "/rtrvr/browser_results.json",
		"final_export":      base + "/exports/final_workflow.json",
	}
}
