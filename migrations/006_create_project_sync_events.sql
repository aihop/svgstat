CREATE TABLE IF NOT EXISTS project_sync_events (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL UNIQUE,
    project_id TEXT,
    external_project_id TEXT,
    tenant_id TEXT,
    sync_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL,
    error_code TEXT,
    error_message TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_status CHECK (status IN ('received', 'processed', 'rejected', 'failed')),
    CONSTRAINT chk_sync_type CHECK (sync_type IN ('project.upsert', 'project.disable', 'project.rotateKey', 'project.refreshCache')),
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_project_sync_events_sync_type ON project_sync_events(sync_type);
CREATE INDEX IF NOT EXISTS idx_project_sync_events_status ON project_sync_events(status);
CREATE INDEX IF NOT EXISTS idx_project_sync_events_external_project_id ON project_sync_events(external_project_id);
CREATE INDEX IF NOT EXISTS idx_project_sync_events_created_at ON project_sync_events(created_at);
