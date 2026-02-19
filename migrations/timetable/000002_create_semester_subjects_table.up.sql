CREATE TABLE IF NOT EXISTS semester_subjects (
    semester_id UUID NOT NULL REFERENCES semesters(id) ON DELETE CASCADE,
    subject_id  UUID NOT NULL,
    teacher_id  UUID,
    CONSTRAINT pk_semester_subjects PRIMARY KEY (semester_id, subject_id)
);

CREATE INDEX IF NOT EXISTS idx_semester_subjects_semester_id ON semester_subjects(semester_id);
CREATE INDEX IF NOT EXISTS idx_semester_subjects_teacher_id  ON semester_subjects(teacher_id);
