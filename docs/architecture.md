# ActR.AI — Architecture Document

## Original GoTube Foundations

GoTube provided the following production-quality patterns that ActR.AI builds on:

- **S3-compatible storage abstraction** — interface + S3 client. Reused verbatim, target swapped from MinIO/R2 to Tigris.
- **Presigned PUT upload flow** — browser uploads directly to object storage. Preserved for zero-copy file ingestion.
- **Upload session + status lifecycle** — extended into the full workflow state machine.
- **Go repository pattern** — type-safe database access layer. Extended to cover workflow jobs, steps, action cards, claims, browser runs, execution runs.
- **Handler helpers** — `decodeJSON`, `writeJSON`, `writeError`. Adapted into `internal/httpx`.
- **CORS handling** — preserved and extended.
- **Health/readiness endpoints** — extended with provider availability checks.
- **Migration-based schema management** — same numbered migration style.

## Browser Research Provider

Browser automation and source research are handled exclusively by **Rtrvr.ai** (`internal/rtrvr`). No other browser research provider is used.

---

## Source Modes

| Mode | Source Type | Rights |
| --- | --- | --- |
| YouTube Link | `youtube` | `youtube_embed_only` |
| File Upload | `upload` | `uploaded_by_user` |
| Authorized Direct URL | `authorized_direct_file` | `authorized_direct_file` |

---

## Tigris Artifact Model

```text
videos/{jobId}/source/original.mp4           ← uploaded or downloaded file
videos/{jobId}/source/transcript.vtt         ← user-uploaded transcript
videos/{jobId}/youtube/metadata.json         ← YouTube API response
videos/{jobId}/transcript/transcript.json    ← parsed transcript
videos/{jobId}/analysis/summary.json         ← AI-generated summary
videos/{jobId}/analysis/action_cards.json    ← AI-generated actions
videos/{jobId}/analysis/claims.json          ← AI-extracted claims
videos/{jobId}/rtrvr/source_research.json    ← Rtrvr source research task
videos/{jobId}/rtrvr/browser_results.json    ← Rtrvr browser results
videos/{jobId}/daytona/execution_logs.json   ← Daytona run output
videos/{jobId}/exports/final_workflow.json   ← Complete workflow export
```

All objects are private. Access is via signed URLs with configurable TTL.

---

## Workflow State Machine

```text
created
  ↓
queued
  ↓
metadata_fetching      (YouTube mode: fetch title, description, embed URL)
  ↓
source_ready           (file available in Tigris or YouTube metadata stored)
  ↓ or → waiting_for_user_input  (transcript needed — user must upload)
transcribing           (prepare transcript text)
  ↓
chunking               (segment for context windows)
  ↓
summarizing            (NVIDIA NIM → summary.json in Tigris)
  ↓
extracting_actions     (NVIDIA NIM → action_cards.json in Tigris, DB rows)
  ↓
extracting_claims      (NVIDIA NIM → claims.json in Tigris, DB rows)
  ↓
researching_with_rtrvr (Rtrvr → browser research, update claims, browser_runs row)
  ↓
finalizing             (assemble final_workflow.json)
  ↓
ready ✓                OR  failed ✗ (at any step)
```

Each step is bounded — it runs within a single Vercel Function invocation.
The frontend polls `GET /api/workflows/{jobId}` every 2 seconds and calls `POST /api/workflows/{jobId}/continue` to advance.

---

## Provider Responsibilities

| Provider | Package | Responsibilities |
| --- | --- | --- |
| Tigris | `internal/storage` | Store/retrieve all artifacts, presigned URLs, signed playback |
| NVIDIA NIM | `internal/nvidia` + `internal/agents` | Summary, action cards, claims extraction, JSON validation + repair, Rtrvr task generation |
| Rtrvr | `internal/rtrvr` | Browser automation, source gathering, claim research, docs lookup |
| Daytona | `internal/daytona` | Create sandbox, run code, capture logs, delete sandbox |
| YouTube | `internal/youtube` | Parse video ID, fetch Data API / oEmbed metadata |
| InsForge/Postgres | `internal/db` | InsForge HTTP adapter + Postgres adapter + memory (dev) |

---

## Frontend Architecture

```text
apps/web/
├── pages/
│   ├── index.vue           Landing + YouTube quick submit
│   ├── new.vue             Three source mode cards
│   └── workflows/[id].vue  Live dashboard with polling
├── composables/
│   ├── useApi.ts           Typed API client + uploadToTigris helper
│   └── useWorkflowPolling.ts  Polling + auto-advance logic
├── stores/
│   └── workflow.ts         Recent jobs (persisted to localStorage)
├── components/
│   ├── layout/AppShell     Nav + footer + sponsor badges
│   ├── workflow/           Timeline, SummaryCard, WaitingForInputPanel
│   ├── actions/            ActionCard, ClaimCard, BrowserResearchPanel, ExecutionLogPanel
│   ├── video/              VideoPreview (YouTube embed + Tigris playback)
│   └── upload/             UploadDropzone, DirectFileUrlForm
```

---

## Security Notes

- Direct file URL validation rejects localhost, private IP ranges, non-HTTPS, disallowed content types.
- Webhook requests are validated with a shared secret header.
- Code execution never happens inside Vercel — only via Daytona sandboxes.
- Browser automation never happens inside Vercel — only via Rtrvr.ai.
- Tigris objects are private; only signed URLs are returned to clients.
- YouTube content is embedded via the standard YouTube iframe API. No arbitrary downloading.
