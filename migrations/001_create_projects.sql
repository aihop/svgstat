CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    external_project_id TEXT NOT NULL UNIQUE,
    tenant_id TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'public',
    public_token_hash TEXT,
    default_theme TEXT,
    default_locale TEXT,
    render_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    badge_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    widget_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    chart_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    last_synced_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_status CHECK (status IN ('pending', 'active', 'grace', 'disabled', 'archived')),
    CONSTRAINT chk_visibility CHECK (visibility IN ('public', 'private'))
);

CREATE INDEX IF NOT EXISTS idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_last_synced_at ON projects(last_synced_at);
