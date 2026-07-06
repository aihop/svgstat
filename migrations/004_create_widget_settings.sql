CREATE TABLE IF NOT EXISTS widget_settings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    widget_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    settings JSONB NOT NULL DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_status CHECK (status IN ('active', 'disabled')),
    CONSTRAINT chk_version CHECK (version >= 1),
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT unique_project_widget UNIQUE (project_id, widget_key)
);

CREATE INDEX IF NOT EXISTS idx_widget_settings_status ON widget_settings(status);
