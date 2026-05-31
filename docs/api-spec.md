# ActR.AI — API Specification

Base URL (production): `https://gorube-api.vercel.app`
Base URL (local): `http://localhost:8080`

All responses are JSON. All errors return `{ "error": "message" }`.

---

## Health

### GET /api/health

No auth required.

**Response 200:**

```json
{ "ok": true, "service": "gorube-api" }
```

### GET /api/ready

Returns provider availability.

**Response 200:**

```json
{
  "ok": true,
  "checks": {
    "storage": "ok",
    "db": "ok",
    "ai": "configured",
    "daytona": "configured",
    "rtrvr": "configured",
    "youtube_api": "not_configured"
  }
}
```

---

## Uploads

### POST /api/uploads/presign

Initiate a direct-to-Tigris upload. Returns a presigned PUT URL.

**Request:**

```json
{
  "filename": "video.mp4",
  "contentType": "video/mp4",
  "fileSize": 52428800,
  "title": "My Video"
}
```

**Response 201:**

```json
{
  "jobId": "uuid",
  "uploadUrl": "https://...",
  "objectKey": "videos/{jobId}/source/original.mp4"
}
```

### POST /api/uploads/transcript/presign

Presigned URL for uploading a transcript/audio fallback.

**Request:**

```json
{
  "jobId": "uuid",
  "filename": "transcript.vtt",
  "contentType": "text/vtt"
}
```

**Response 200:**

```json
{
  "uploadUrl": "https://...",
  "objectKey": "videos/{jobId}/source/transcript.vtt"
}
```

---

## Workflows

### POST /api/workflows/youtube/start

**Request:**

```json
{ "youtubeUrl": "https://www.youtube.com/watch?v=dQw4w9WgXcQ" }
```

**Response 201:**

```json
{
  "jobId": "uuid",
  "status": "waiting_for_user_input",
  "requiresUpload": true,
  "video": {
    "youtubeVideoId": "dQw4w9WgXcQ",
    "youtubeEmbedUrl": "https://www.youtube.com/embed/dQw4w9WgXcQ",
    "title": "Never Gonna Give You Up",
    "thumbnailUrl": "https://...",
    "channel": "Rick Astley"
  },
  "message": "YouTube metadata fetched. Please upload a transcript to continue."
}
```

### POST /api/workflows/upload/start

**Request:**

```json
{
  "jobId": "uuid",
  "objectKey": "videos/{jobId}/source/original.mp4",
  "title": "My Video"
}
```

**Response 200:**

```json
{ "jobId": "uuid", "status": "transcribing" }
```

### POST /api/workflows/direct-file/start

**Request:**

```json
{
  "fileUrl": "https://example.com/lecture.mp4",
  "title": "Lecture Recording",
  "sourceRightsConfirmed": true
}
```

**Response 201:**

```json
{ "jobId": "uuid", "status": "source_ready" }
```

### GET /api/workflows/{jobId}

Get complete workflow details including steps, action cards, claims, browser runs, execution runs, and signed artifact URLs.

**Response 200:** `WorkflowDetails` — see `apps/web/types/index.ts`.

### POST /api/workflows/{jobId}/continue

Advance the workflow by one bounded step.

**Response 200:**

```json
{
  "jobId": "uuid",
  "status": "summarizing",
  "progress": 40,
  "requiresUserInput": false
}
```

---

## Actions

### POST /api/actions/{actionCardId}/run

Run a code action card via Daytona.

**Request:**

```json
{ "mode": "daytona" }
```

**Response 200:**

```json
{
  "runId": "uuid",
  "status": "completed",
  "exitCode": 0,
  "logsKey": "videos/{jobId}/daytona/execution_logs.json"
}
```

If Daytona is not configured:

```json
{
  "error": "provider_not_configured",
  "provider": "Daytona",
  "detail": "Provider not configured: add DAYTONA_API_KEY to enable Daytona."
}
```

---

## Browser Research

### POST /api/rtrvr/run

Start a Rtrvr browser automation / source research task.

**Request:**

```json
{
  "jobId": "uuid",
  "actionCardId": "uuid (optional)",
  "task": "Find official documentation and sources for these claims",
  "targetUrls": ["https://example.com"]
}
```

**Response 200:**

```json
{
  "browserRunId": "uuid",
  "status": "completed",
  "outputKey": "videos/{jobId}/rtrvr/browser_results.json",
  "taskId": "rtrvr-task-id"
}
```

If Rtrvr is not configured:

```json
{
  "error": "provider_not_configured",
  "provider": "Rtrvr",
  "detail": "Provider not configured: add RTRVR_API_KEY to enable browser research."
}
```

---

## Webhooks

### POST /api/webhooks/tigris

Called by Tigris when an object is created. Advances job status for matching source keys.

Validates `X-Webhook-Secret` header if `WEBHOOK_SECRET` is set.

---

## Error Codes

| HTTP | Meaning |
| --- | --- |
| 400 | Bad request / invalid JSON |
| 422 | Validation error (field-level) |
| 404 | Not found |
| 503 | Provider not configured |
| 502 | External provider error |
