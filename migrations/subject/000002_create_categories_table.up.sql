-- subject_categories table for grouping subjects (schema-per-tenant)
CREATE TABLE IF NOT EXISTS subject_categories (
    id          UUID PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add foreign key from subjects to categories now that both tables exist.
ALTER TABLE subjects
    ADD CONSTRAINT subjects_category_id_fk
    FOREIGN KEY (category_id) REFERENCES subject_categories (id) ON DELETE SET NULL;
