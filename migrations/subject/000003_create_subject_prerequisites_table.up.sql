-- subject_prerequisites stores directed prerequisite edges (schema-per-tenant).
-- Edge (subject_id → prerequisite_id) means: subject_id requires prerequisite_id.
CREATE TABLE IF NOT EXISTS subject_prerequisites (
    subject_id      UUID NOT NULL REFERENCES subjects (id) ON DELETE CASCADE,
    prerequisite_id UUID NOT NULL REFERENCES subjects (id) ON DELETE CASCADE,
    PRIMARY KEY (subject_id, prerequisite_id)
);

-- subject_prerequisite_versions holds the optimistic-locking version counter
-- per subject. Incremented on every AddEdge / RemoveEdge for that subject.
CREATE TABLE IF NOT EXISTS subject_prerequisite_versions (
    subject_id UUID PRIMARY KEY REFERENCES subjects (id) ON DELETE CASCADE,
    version    INT NOT NULL DEFAULT 1
);
