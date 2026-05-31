-- Migration 001: ActR.AI initial schema
-- Target: postgresql://...@hhm46uz6.us-east.database.insforge.app:5432/insforge

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT UNIQUE NOT NULL,
    name        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_jobs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID REFERENCES users(id) ON DELETE SET NULL,
    source_type             TEXT NOT NULL CHECK (source_type IN ('youtube', 'upload', 'authorized_direct_file')),
    source_rights           TEXT NOT NULL CHECK (source_rights IN ('youtube_embed_only', 'uploaded_by_user', 'authorized_direct_file')),
    source_url              TEXT,
    source_tigris_key       TEXT,
    transcript_tigris_key   TEXT,
    title                   TEXT,
    status                  TEXT NOT NULL DEFAULT 'created',
    current_step            TEXT,
    progress                INT NOT NULL DEFAULT 0,
    requires_user_input     BOOLEAN NOT NULL DEFAULT FALSE,
    required_input_type     TEXT,
    error_message           TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflow_jobs_status     ON workflow_jobs(status);
CREATE INDEX IF NOT EXISTS idx_workflow_jobs_created_at ON workflow_jobs(created_at DESC);

CREATE TABLE IF NOT EXISTS videos (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id              UUID NOT NULL REFERENCES workflow_jobs(id) ON DELETE CASCADE,
    title               TEXT NOT NULL,
    description         TEXT,
    duration_seconds    INT,
    thumbnail_url       TEXT,
    youtube_video_id    TEXT,
    youtube_embed_url   TEXT,
    tigris_video_key    TEXT,
    transcript_key      TEXT,
    playback_url        TEXT,
    status              TEXT NOT NULL DEFAULT 'created',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflow_steps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES workflow_jobs(id) ON DELETE CASCADE,
    step_name       TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    output_key      TEXT,
    error_message   TEXT,
    UNIQUE (job_id, step_name)
);

CREATE TABLE IF NOT EXISTS action_cards (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id              UUID NOT NULL REFERENCES workflow_jobs(id) ON DELETE CASCADE,
    title               TEXT NOT NULL,
    description         TEXT NOT NULL,
    action_type         TEXT NOT NULL CHECK (action_type IN ('checklist', 'code', 'research', 'study_plan', 'browser_action')),
    timestamp_seconds   INT,
    status              TEXT NOT NULL DEFAULT 'pending',
    provider            TEXT,
    requires_execution  BOOLEAN NOT NULL DEFAULT FALSE,
    output_key          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS claims (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id                  UUID NOT NULL REFERENCES workflow_jobs(id) ON DELETE CASCADE,
    claim_text              TEXT NOT NULL,
    timestamp_seconds       INT,
    verification_status     TEXT NOT NULL DEFAULT 'pending',
    confidence              NUMERIC(5,4),
    evidence_key            TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS execution_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_card_id  UUID NOT NULL REFERENCES action_cards(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    input_key       TEXT,
    output_key      TEXT,
    logs_key        TEXT,
    exit_code       INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS browser_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES workflow_jobs(id) ON DELETE CASCADE,
    action_card_id  UUID REFERENCES action_cards(id) ON DELETE SET NULL,
    provider        TEXT NOT NULL DEFAULT 'rtrvr',
    status          TEXT NOT NULL DEFAULT 'pending',
    task            TEXT NOT NULL,
    input_key       TEXT,
    output_key      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
