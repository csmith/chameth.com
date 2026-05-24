CREATE TABLE wow_achievements (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    achievement_id INTEGER NOT NULL,
    achievement_name TEXT NOT NULL,
    completed_at TIMESTAMP NOT NULL,
    character_id INTEGER NOT NULL REFERENCES wow_characters(id) ON DELETE CASCADE,
    UNIQUE (achievement_id, completed_at, character_id)
);
