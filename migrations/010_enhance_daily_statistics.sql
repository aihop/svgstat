ALTER TABLE daily_statistics
ADD COLUMN IF NOT EXISTS paths JSONB NOT NULL DEFAULT '{}',
ADD COLUMN IF NOT EXISTS ips JSONB NOT NULL DEFAULT '{}';

COMMENT ON COLUMN daily_statistics.paths IS 'Page paths accessed';
COMMENT ON COLUMN daily_statistics.ips IS 'Unique IP addresses';
