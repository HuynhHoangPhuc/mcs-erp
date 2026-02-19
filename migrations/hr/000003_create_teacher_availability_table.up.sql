CREATE TABLE IF NOT EXISTS teacher_availability (
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    day INTEGER NOT NULL CHECK (day >= 0 AND day <= 6),
    period INTEGER NOT NULL CHECK (period >= 1 AND period <= 10),
    is_available BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT uq_teacher_day_period UNIQUE (teacher_id, day, period)
);

CREATE INDEX IF NOT EXISTS idx_teacher_availability_teacher_id ON teacher_availability(teacher_id);
