CREATE TABLE IF NOT EXISTS project_limits (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL UNIQUE,
    plan_code TEXT NOT NULL,
    max_requests_per_minute INTEGER,
    cache_ttl_seconds INTEGER,
    badge_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    widget_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    chart_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    render_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    effective_from TIMESTAMPTZ,
    effective_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_project_limits_plan_code ON project_limits(plan_code);
CREATE INDEX IF NOT EXISTS idx_project_limits_effective_until ON project_limits(effective_until);
