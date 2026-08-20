-- Migration 001_init.sql
-- Initialises the schema for the welfare settlement exception resolution platform.
-- Migrations are idempotent: re-running them is a no-op.

CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    sponsor TEXT NOT NULL,
    annual_budget_cents BIGINT NOT NULL DEFAULT 0,
    start_year INTEGER NOT NULL,
    end_year INTEGER NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS parties (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('donor','implementer','beneficiary','intermediary')),
    contact TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS funding_batches (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    code TEXT NOT NULL,
    donor_party_id TEXT NOT NULL,
    implementer_party_id TEXT NOT NULL,
    intermediary_party_id TEXT,
    total_amount_cents BIGINT NOT NULL CHECK (total_amount_cents > 0),
    currency TEXT NOT NULL,
    disbursed_at TIMESTAMPTZ NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS settlement_cycles (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    funding_batch_id TEXT NOT NULL REFERENCES funding_batches(id) ON DELETE RESTRICT,
    year INTEGER NOT NULL CHECK (year > 0),
    quarter INTEGER NOT NULL CHECK (quarter BETWEEN 1 AND 4),
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_cycle_dates CHECK (end_date >= start_date)
);

CREATE TABLE IF NOT EXISTS rule_versions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    code TEXT NOT NULL,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    description TEXT NOT NULL DEFAULT '',
    rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    UNIQUE (tenant_id, code)
);

CREATE TABLE IF NOT EXISTS settlement_entries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    cycle_id TEXT NOT NULL REFERENCES settlement_cycles(id) ON DELETE RESTRICT,
    batch_id TEXT NOT NULL REFERENCES funding_batches(id) ON DELETE RESTRICT,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    source_id TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('import','manual','resubmit')),
    payee_party_id TEXT NOT NULL,
    payer_party_id TEXT NOT NULL,
    amount_cents BIGINT NOT NULL DEFAULT 0 CHECK (amount_cents >= 0),
    currency TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    source_fingerprint TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_fingerprint)
);

CREATE INDEX IF NOT EXISTS idx_entries_cycle_id ON settlement_entries (cycle_id);
CREATE INDEX IF NOT EXISTS idx_entries_project_id ON settlement_entries (project_id);
CREATE INDEX IF NOT EXISTS idx_entries_source_id ON settlement_entries (source_id);

CREATE TABLE IF NOT EXISTS exceptions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    cycle_id TEXT NOT NULL REFERENCES settlement_cycles(id) ON DELETE RESTRICT,
    entry_id TEXT NOT NULL REFERENCES settlement_entries(id) ON DELETE RESTRICT,
    rule_version_id TEXT NOT NULL REFERENCES rule_versions(id) ON DELETE RESTRICT,
    rule_code TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('low','medium','high','critical')),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    hit_reason TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','review','resolved','closed','escalated')),
    assignee_id TEXT,
    reporter_id TEXT,
    deadline_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (entry_id, rule_code)
);

CREATE INDEX IF NOT EXISTS idx_exceptions_cycle_id ON exceptions (cycle_id);
CREATE INDEX IF NOT EXISTS idx_exceptions_assignee_id ON exceptions (assignee_id);
CREATE INDEX IF NOT EXISTS idx_exceptions_status ON exceptions (status);

CREATE TABLE IF NOT EXISTS exception_notes (
    id TEXT PRIMARY KEY,
    exception_id TEXT NOT NULL REFERENCES exceptions(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL,
    body TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('comment','assignment','claim','resubmit','review','escalation','rework')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exception_attachments (
    id TEXT PRIMARY KEY,
    exception_id TEXT NOT NULL REFERENCES exceptions(id) ON DELETE CASCADE,
    original_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    sha256 TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    uploaded_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sha256)
);

CREATE TABLE IF NOT EXISTS summaries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    cycle_id TEXT NOT NULL REFERENCES settlement_cycles(id) ON DELETE RESTRICT,
    rule_version_id TEXT NOT NULL,
    computed_at TIMESTAMPTZ NOT NULL,
    total_entries INTEGER NOT NULL DEFAULT 0,
    total_amount_cents BIGINT NOT NULL DEFAULT 0,
    approved_amount_cents BIGINT NOT NULL DEFAULT 0,
    pending_amount_cents BIGINT NOT NULL DEFAULT 0,
    exception_count_by_status JSONB NOT NULL DEFAULT '{}'::jsonb,
    exception_count_by_severity JSONB NOT NULL DEFAULT '{}'::jsonb,
    diff_basis JSONB NOT NULL DEFAULT '{}'::jsonb,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_summaries_cycle_version ON summaries (cycle_id, version DESC);

CREATE TABLE IF NOT EXISTS recalculation_batches (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    cycle_id TEXT NOT NULL,
    rule_version_id TEXT NOT NULL,
    input_digest TEXT NOT NULL,
    trigger_reason TEXT NOT NULL,
    trigger_rule_code TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running','completed','failed')),
    output_summary JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS annual_accumulators (
    project_id TEXT NOT NULL,
    year INTEGER NOT NULL,
    budget_cents BIGINT NOT NULL DEFAULT 0,
    disbursed_cents BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (project_id, year)
);

CREATE TABLE IF NOT EXISTS annual_adjustments (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    year INTEGER NOT NULL,
    delta_cents BIGINT NOT NULL,
    reason TEXT NOT NULL,
    triggered_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_annual_adjustments_project_year ON annual_adjustments (project_id, year);

CREATE TABLE IF NOT EXISTS audit_entries (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    detail JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_entries (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_entries (actor_id);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    username TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL CHECK (role IN ('operator','assignee','reviewer','admin')),
    password_hash TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, username)
);
