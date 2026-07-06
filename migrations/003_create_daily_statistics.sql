CREATE TABLE IF NOT EXISTS daily_statistics (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    date DATE NOT NULL,
    pv BIGINT NOT NULL DEFAULT 0,
    uv BIGINT NOT NULL DEFAULT 0,
    requests BIGINT NOT NULL DEFAULT 0,
    bots BIGINT NOT NULL DEFAULT 0,
    referrers JSONB NOT NULL DEFAULT '{}',
    countries JSONB NOT NULL DEFAULT '{}',
    devices JSONB NOT NULL DEFAULT '{}',
    browsers JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT unique_project_date UNIQUE (project_id, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_statistics_date ON daily_statistics(date);
