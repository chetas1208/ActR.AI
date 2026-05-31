# Gorube Flow

**Turn videos into workflows that act.**

Gorube Flow is a video-to-agent workflow engine built on [GoTube](https://github.com/chetas1208/GoTube). Paste a YouTube link, upload media, or provide an authorized direct file URL. Gorube Flow extracts what matters, gathers sources with Rtrvr browser automation, generates action cards, and safely executes code steps via Daytona.

---

## Built on GoTube

Gorube Flow preserves and extends the best architectural patterns from GoTube:

| GoTube Pattern | Gorube Flow Usage |
| --- | --- |
| S3-compatible storage abstraction | Reused for Tigris (same interface) |
| Presigned PUT URL upload flow | Preserved for direct-to-Tigris uploads |
| Video status lifecycle model | Extended into `workflow_jobs` state machine |
| Upload initiation + completion | Adapted into presign + start endpoints |
| Repository pattern | Extended for workflow, action cards, claims, browser runs, execution runs |
| Handler helpers (decodeJSON, writeJSON) | Adapted into `internal/httpx` package |
| CORS handling | Preserved with `httpx.ApplyCORS` |
| Health/readiness endpoints | Extended with provider availability checks |
| Migration pattern | Preserved in `apps/api/migrations/` |

---

## Architecture

```text
gorube/
├── apps/
│   ├── web/          Nuxt 4 + TypeScript + Tailwind — deployed as gorube-web
│   └── api/          Go Vercel Functions — deployed as gorube-api
├── infra/
│   ├── db/           schema.sql (Postgres-compatible)
│   └── env/          *.env.example files
└── docs/             API spec, architecture, workflow lifecycle
```

### System Flow

```text
User Input
    │
    ├─ YouTube URL ──► YouTube Data API / oEmbed → metadata in Tigris
    │                  → If transcript unavailable: waiting_for_user_input
    │
    ├─ File Upload ──► Browser → Presigned PUT → Tigris → workflow starts
    │
    └─ Direct URL ──► Validation → Stream download → Tigris → workflow starts
                                                              │
                                                              ▼
                              Bounded workflow steps (frontend polls + /continue):
                              1. prepare_transcript
                              2. chunk_transcript
                              3. generate_summary      (OpenAI)
                              4. extract_action_cards  (OpenAI)
                              5. extract_claims        (OpenAI)
                              6. rtrvr_research        (Rtrvr browser automation)
                              7. finalize → all artifacts in Tigris
```

---

## Provider Integrations

| Provider | Role | Required Env Var |
| --- | --- | --- |
| **Tigris** | Object storage for all artifacts, presigned uploads, signed playback | `TIGRIS_BUCKET`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| **Daytona** | Safe isolated code execution | `DAYTONA_API_KEY` |
| **InsForge** | Hosted Postgres database | `DATABASE_URL` |
| **Rtrvr.ai** | Browser automation, source research, claims assessment | `RTRVR_API_KEY` |
| **OpenAI-compatible** | Summaries, action cards, claims extraction | `OPENAI_API_KEY` |
| **YouTube Data API** | Video metadata (optional) | `YOUTUBE_API_KEY` |
| **Vercel** | Deployment (web + API) | — |

---

## Source Mode Policy

### 1. `youtube_link`

- Paste any YouTube URL.
- Backend parses the video ID and fetches metadata via YouTube Data API or oEmbed.
- Frontend embeds the YouTube player.
- Transcript is **not** downloaded arbitrarily. If captions/transcript are unavailable via permitted paths, the workflow pauses and requests user upload.
- **No arbitrary YouTube video downloading is implemented.**

### 2. `uploaded_video`

- User uploads a local video, audio, or transcript file.
- Browser uploads directly to Tigris via presigned PUT URL (GoTube pattern preserved).
- Full workflow runs on the uploaded file.

### 3. `authorized_direct_file_url`

- User provides an HTTPS URL to a file they own or have permission to process.
- Backend validates: HTTPS only, no private IPs, allowed content types, max size, redirect safety.
- File is streamed directly into Tigris.
- Same workflow pipeline runs as for uploaded files.

---

## Local Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- pnpm 9+

### 1. Clone and install

```bash
git clone https://github.com/chetas1208/GoTube.git gorube-flow
cd gorube-flow
pnpm install
```

### 2. Configure environment

```bash
cp .env /path/to/.env.local
# or copy infra/env/api.env.example → apps/api/.env.local
# and infra/env/web.env.example → apps/web/.env.local
```

### 3. Run DB migration

```bash
psql $DATABASE_URL < apps/api/migrations/001_gorube_flow_schema.up.sql
```

### 4. Start development

```bash
pnpm dev:web             # Nuxt on :3000
cd apps/api && vercel dev --listen 8080   # API functions on :8080
```

---

## Environment Variables

See `.env` for the full list with real credentials (local dev only — do not commit).

See `infra/env/api.env.example` for the production Vercel project reference.

### Required for production

| Variable | Description |
| --- | --- |
| `TIGRIS_BUCKET` | Tigris bucket name |
| `TIGRIS_ENDPOINT` | Tigris endpoint (`https://t3.storage.dev`) |
| `AWS_ACCESS_KEY_ID` | Tigris access key ID |
| `AWS_SECRET_ACCESS_KEY` | Tigris secret access key |
| `DATABASE_URL` | Postgres connection string |
| `OPENAI_API_KEY` | OpenAI-compatible key |
| `DAYTONA_API_KEY` | Daytona key |
| `DAYTONA_API_URL` | Daytona API URL |
| `RTRVR_API_KEY` | Rtrvr.ai key |

---

## Vercel Deployment

### gorube-api

1. Create Vercel project, connect repo
2. Root Directory: `apps/api`
3. Framework: Other
4. Add all API env vars
5. Deploy

### gorube-web

1. Create second Vercel project, connect same repo
2. Root Directory: `apps/web`
3. Framework: Nuxt.js
4. Add `NUXT_PUBLIC_API_BASE_URL` pointing to gorube-api URL
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
| POST | `/api/workflows/direct-file/start` | Start direct file URL workflow |
| GET | `/api/workflows/{jobId}` | Get full workflow details |
| POST | `/api/workflows/{jobId}/continue` | Advance to next bounded step |
| POST | `/api/actions/{actionCardId}/run` | Run action card via Daytona |
| POST | `/api/rtrvr/run` | Run Rtrvr browser research task |
| POST | `/api/webhooks/tigris` | Tigris object-created webhook |

---

## Workflow State Machine

```text
created → queued → metadata_fetching → source_ready
       → waiting_for_user_input (if transcript needed)
       → transcribing → chunking → summarizing
       → extracting_actions → extracting_claims
       → researching_with_rtrvr
       → finalizing → ready
                    └→ failed (at any step)
```

---

## Known Limitations

- YouTube transcript is not automatically downloaded — the app requests user upload when captions are unavailable.
- Large video transcription is not implemented natively (no FFmpeg in Vercel) — use the transcript upload fallback.
- Vercel Function timeout limits per-step work. The frontend polls and advances step by step.
- The in-memory DB adapter does not persist across Vercel cold starts — use Postgres (`DATABASE_URL`) in production.
- Missing provider keys produce honest "provider not configured" messages — never mocked results.

---

## License

MIT — see [LICENSE](./LICENSE)
