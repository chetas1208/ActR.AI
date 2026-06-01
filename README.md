# ActR.AI

**Video to agent workflow engine.**

ActR.AI turns video into action. Paste a YouTube link, upload media, or provide an authorized direct file URL. ActR.AI extracts what matters, generates action cards, researches sources through Rtrvr.ai browser automation, and safely executes code steps through Daytona — powered by NVIDIA NIM.

---

## Built on GoTube

ActR.AI preserves and extends the best architectural patterns from [GoTube](https://github.com/chetas1208/GoTube):

| GoTube Pattern | ActR.AI Usage |
| --- | --- |
| S3-compatible storage abstraction | Reused for Tigris |
| Presigned PUT upload flow | Preserved for direct-to-Tigris uploads |
| Upload status lifecycle | Extended into the full workflow state machine |
| Go repository pattern | Extended for workflows, action cards, claims, browser runs, execution runs |
| Handler helpers (decodeJSON, writeJSON) | Adapted into `internal/httpx` |
| CORS + health endpoints | Preserved and extended with provider checks |
| Migration-based schema | Same numbered migration style |

---

## Architecture

```text
actr-ai/
├── apps/
│   ├── web/    Nuxt 4 + TypeScript + Tailwind — deployed as actr-ai-web
│   └── api/    Go Vercel Functions — deployed as actr-ai-api
├── infra/
│   ├── db/     schema.sql (Postgres-compatible, includes browser_runs)
│   └── env/    api.env.example, web.env.example
└── docs/       api-spec.md, architecture.md
```

### Workflow Pipeline

```text
User Input
    │
    ├─ YouTube URL ──► YouTube Data API → metadata + embed
    │                  → If no transcript: waiting_for_user_input
    │
    ├─ File Upload ──► Browser → Presigned PUT → Tigris
    │
    └─ Direct URL ──► Validated download → Tigris
                                           │
                                           ▼
                 Bounded steps (frontend polls + /continue):
                 1. prepare_transcript   (NVIDIA ASR or upload)
                 2. chunk_transcript
                 3. generate_summary     (NVIDIA NIM)
                 4. extract_action_cards (NVIDIA NIM)
                 5. extract_claims       (NVIDIA NIM)
                 6. rtrvr_research       (Rtrvr browser automation)
                 7. finalize → Tigris export
```

---

## NVIDIA NIM Model Strategy

| Model | Purpose | Env Var |
| --- | --- | --- |
| `nvidia/nemotron-3-super-120b-a12b` | Primary reasoning, action extraction, claims, summaries | `NIM_LLM_MODEL` |
| `nvidia/llama-3.1-nemotron-nano-8b-v1` | Fast tasks: JSON repair, classification | `NIM_FAST_MODEL` |
| configurable | Code execution planning | `NIM_CODING_MODEL` |

All LLM calls use NVIDIA NIM's OpenAI-compatible `/chat/completions` endpoint. No OpenAI API keys are used.

---

## NVIDIA ASR Transcription Strategy

1. **NVIDIA ASR** (preferred) — `NVIDIA_ASR_MODEL=nvidia/parakeet-1.1b-rnnt-multilingual-asr`
   - Configured via `NIM_API_KEY` (or `NVIDIA_ASR_API_KEY` if separate)
   - Supports English and multilingual transcription
   - Stores transcript JSON + VTT in Tigris

2. **Transcript file upload** (fallback) — when ASR is not configured
   - Accepts `.vtt`, `.srt`, `.txt`, `.json`
   - Workflow pauses at `waiting_for_user_input` and resumes after upload

3. **Local Whisper** (dev/self-hosted only)
   - `LOCAL_TRANSCRIPTION_ENABLED=true` + `LOCAL_WHISPER_SERVER_URL`
   - Not allowed on Vercel — returns `provider_not_configured` if attempted

---

## Provider Integrations

| Provider | Role | Env Var |
| --- | --- | --- |
| **NVIDIA NIM** | GenAI reasoning, summaries, action cards, claims | `NIM_API_KEY` |
| **Tigris** | Object storage for all artifacts | `TIGRIS_BUCKET`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| **Daytona** | Safe isolated code execution | `DAYTONA_API_KEY` |
| **InsForge/Postgres** | Hosted database | `DATABASE_URL` |
| **Rtrvr.ai** | Browser automation + source research | `RTRVR_API_KEY` |
| **YouTube Data API** | Video metadata (optional) | `YOUTUBE_API_KEY` |
| **Vercel** | Deployment | — |

---

## Source Mode Policy

### 1. `youtube_link`

- Parses the YouTube video ID from the URL.
- Fetches metadata via YouTube Data API (or oEmbed fallback).
- Embeds the YouTube player in the frontend.
- No arbitrary YouTube video downloading.
- If captions/transcript are unavailable: pauses at `waiting_for_user_input`.

### 2. `uploaded_video`

- User uploads a file; browser writes directly to Tigris via presigned PUT URL.
- NVIDIA ASR transcribes audio/video if configured.
- Transcript files (`.vtt`, `.srt`, `.txt`, `.json`) continue immediately.

### 3. `authorized_direct_file_url`

- HTTPS-only. Rejects private IPs, localhost, non-HTTPS redirects.
- Validates content type and size against `MAX_DIRECT_FILE_BYTES`.
- Streams file directly into Tigris, then starts the same workflow pipeline.

---

## Local Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- pnpm 9+

### 1. Clone and install

```bash
git clone https://github.com/chetas1208/GoTube.git actr-ai
cd actr-ai
pnpm install
```

### 2. Configure environment

```bash
cp infra/env/api.env.example apps/api/.env
cp infra/env/web.env.example apps/web/.env
```

Add your `NIM_API_KEY` from [build.nvidia.com](https://build.nvidia.com/).

### 3. Run DB migration

```bash
psql $DATABASE_URL < apps/api/migrations/001_gorube_flow_schema.up.sql
```

### 4. Start development

```bash
pnpm dev:api                                    # Go API on :8081 (local, no Vercel CLI)
pnpm dev:web                                    # Nuxt on :3000
```

---

## Vercel Deployment

### actr-ai-api

1. Create Vercel project, connect this repo
2. Root Directory: `apps/api`
3. Framework: Other (Go Functions)
4. Add all API env vars from `infra/env/api.env.example`
5. Deploy

### actr-ai-web

1. Create second Vercel project, same repo
2. Root Directory: `apps/web`
3. Framework: Nuxt.js
4. Add `NUXT_PUBLIC_API_BASE_URL` pointing to your `actr-ai-api` URL
5. Deploy

---

## API Routes

| Method | Path | Description |
| --- | --- | --- |
| GET | `/api/health` | Health check |
| GET | `/api/ready` | Readiness + provider status |
| POST | `/api/uploads/presign` | Presigned Tigris upload URL |
| POST | `/api/uploads/transcript/presign` | Presigned transcript upload URL |
| POST | `/api/workflows/youtube/start` | Start YouTube workflow |
| POST | `/api/workflows/upload/start` | Complete upload + start workflow |
| POST | `/api/workflows/direct-file/start` | Start direct URL workflow |
| GET | `/api/workflows/{jobId}` | Full workflow details |
| POST | `/api/workflows/{jobId}/continue` | Advance one bounded step |
| POST | `/api/actions/{actionCardId}/run` | Run action via Daytona |
| POST | `/api/rtrvr/run` | Run Rtrvr browser research |
| POST | `/api/webhooks/tigris` | Tigris object-created webhook |

---

## Workflow State Machine

```text
created → queued → metadata_fetching → source_ready
       → waiting_for_user_input (transcript needed)
       → transcribing → chunking → summarizing
       → extracting_actions → extracting_claims
       → researching_with_rtrvr
       → finalizing → ready
                    └→ failed (at any step)
```

---

## Known Limitations

- YouTube transcript is not automatically downloaded — the app requests user upload when captions are unavailable.
- NVIDIA ASR runs as a bounded Vercel function call. For large files (>10 min audio), the 10-second Vercel timeout may apply — use transcript upload for long recordings.
- The in-memory DB adapter does not persist across Vercel cold starts — use Postgres (`DATABASE_URL`) in production.
- Missing provider keys produce honest "provider not configured" messages — never mocked results.

---

## License

MIT — see [LICENSE](./LICENSE)
