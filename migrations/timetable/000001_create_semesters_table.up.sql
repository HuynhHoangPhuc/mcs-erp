CREATE TABLE IF NOT EXISTS semesters (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    start_date TIMESTAMPTZ  NOT NULL,
    end_date   TIMESTAMPTZ  NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'draft'
                   CHECK (status IN ('draft','scheduling','review','approved','rejected')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_semester_dates CHECK (end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_semesters_status ON semesters(status);
