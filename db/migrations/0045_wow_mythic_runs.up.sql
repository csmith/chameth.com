CREATE TABLE wow_mythic_runs (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES wow_characters(id) ON DELETE CASCADE,
    season_id INTEGER NOT NULL,
    dungeon_id INTEGER NOT NULL,
    dungeon_name TEXT NOT NULL,
    completed_timestamp BIGINT NOT NULL,
    duration BIGINT NOT NULL,
    keystone_level INTEGER NOT NULL,
    is_completed_within_time BOOLEAN NOT NULL,
    mythic_rating DOUBLE PRECISION NOT NULL DEFAULT 0,
    UNIQUE (character_id, season_id, dungeon_id)
);
