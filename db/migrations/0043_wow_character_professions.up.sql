CREATE TABLE wow_character_professions (
    character_id INTEGER NOT NULL REFERENCES wow_characters(id) ON DELETE CASCADE,
    tier_id INTEGER NOT NULL,
    tier_name TEXT NOT NULL,
    profession_id INTEGER NOT NULL,
    profession_name TEXT NOT NULL,
    skill_points INTEGER NOT NULL,
    max_skill_points INTEGER NOT NULL,
    kind TEXT NOT NULL,
    PRIMARY KEY (character_id, tier_id)
);
