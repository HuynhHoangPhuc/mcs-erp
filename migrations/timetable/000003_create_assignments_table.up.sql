CREATE TABLE IF NOT EXISTS schedules (
    semester_id     UUID    NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
    version         INTEGER NOT NULL,
    hard_violations INTEGER NOT NULL DEFAULT 0,
    soft_penalty    NUMERIC NOT NULL DEFAULT 0,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_schedules PRIMARY KEY (semester_id, version)
);

CREATE TABLE IF NOT EXISTS assignments (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    semester_id UUID    NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
    subject_id  UUID    NOT NULL,
    teacher_id  UUID    NOT NULL,
    room_id     UUID    NOT NULL,
    day         INTEGER NOT NULL CHECK (day >= 0 AND day <= 5),
    period      INTEGER NOT NULL CHECK (period >= 1 AND period <= 10),
    version     INTEGER NOT NULL,
    CONSTRAINT fk_assignments_schedule
        FOREIGN KEY (semester_id, version) REFERENCES schedules(semester_id, version) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assignments_semester_version ON assignments(semester_id, version);
CREATE INDEX IF NOT EXISTS idx_assignments_teacher_id       ON assignments(teacher_id);
CREATE INDEX IF NOT EXISTS idx_assignments_room_id          ON assignments(room_id);
CREATE INDEX IF NOT EXISTS idx_assignments_day_period       ON assignments(day, period);
