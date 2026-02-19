CREATE TABLE IF NOT EXISTS rooms (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(255) NOT NULL,
    code      VARCHAR(50)  NOT NULL,
    building  VARCHAR(255) NOT NULL DEFAULT '',
    floor     INTEGER      NOT NULL DEFAULT 0,
    capacity  INTEGER      NOT NULL CHECK (capacity > 0),
    equipment TEXT[]       NOT NULL DEFAULT '{}',
    is_active BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rooms_code_unique UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_rooms_building ON rooms(building);
CREATE INDEX IF NOT EXISTS idx_rooms_capacity ON rooms(capacity);
