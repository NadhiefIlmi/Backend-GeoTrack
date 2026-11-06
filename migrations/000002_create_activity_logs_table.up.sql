CREATE TABLE IF NOT EXISTS activity_logs (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    entity_id VARCHAR(36) NOT NULL,
    entity_code VARCHAR(30) NOT NULL,
    entity_name VARCHAR(100) NOT NULL,
    action VARCHAR(30) NOT NULL,
    details TEXT NOT NULL,
    actor VARCHAR(50) NOT NULL DEFAULT 'System',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_activity_action CHECK (action IN ('created', 'updated', 'status_changed', 'deleted')),
    CONSTRAINT fk_activity_entity FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_activity_entity_id ON activity_logs (entity_id);
CREATE INDEX IF NOT EXISTS idx_activity_action ON activity_logs (action);
CREATE INDEX IF NOT EXISTS idx_activity_created_at ON activity_logs (created_at DESC);

