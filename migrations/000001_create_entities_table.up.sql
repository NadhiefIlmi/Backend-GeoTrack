CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS entities (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4()::text,
    code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_entity_type CHECK (type IN ('vehicle', 'iot_device', 'facility', 'infrastructure')),
    CONSTRAINT chk_entity_status CHECK (status IN ('active', 'inactive', 'maintenance')),
    CONSTRAINT chk_latitude CHECK (latitude >= -90.0 AND latitude <= 90.0),
    CONSTRAINT chk_longitude CHECK (longitude >= -180.0 AND longitude <= 180.0)
);

CREATE INDEX IF NOT EXISTS idx_entities_code ON entities (code);
CREATE INDEX IF NOT EXISTS idx_entities_type ON entities (type);
CREATE INDEX IF NOT EXISTS idx_entities_status ON entities (status);
CREATE INDEX IF NOT EXISTS idx_entities_coordinates ON entities (latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_entities_created_at ON entities (created_at DESC);

