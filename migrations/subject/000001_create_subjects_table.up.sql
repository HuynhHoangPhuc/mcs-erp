-- subjects table for the subject module (schema-per-tenant)
CREATE TABLE IF NOT EXISTS subjects (
    id            UUID PRIMARY KEY,
    name          TEXT        NOT NULL,
    code          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    category_id   UUID,
    credits       INT         NOT NULL DEFAULT 0,
    hours_per_week INT        NOT NULL DEFAULT 0,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT subjects_code_unique UNIQUE (code)
);
