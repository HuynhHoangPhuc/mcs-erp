CREATE TABLE IF NOT EXISTS room_availability (
    room_id      UUID     NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    day          SMALLINT NOT NULL CHECK (day >= 0 AND day <= 6),
    period       SMALLINT NOT NULL CHECK (period >= 1 AND period <= 10),
    is_available BOOLEAN  NOT NULL DEFAULT true,
    CONSTRAINT room_availability_unique UNIQUE (room_id, day, period)
);

CREATE INDEX IF NOT EXISTS idx_room_availability_room_id ON room_availability(room_id);
