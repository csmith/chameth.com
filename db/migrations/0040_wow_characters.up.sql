CREATE TABLE wow_characters (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    character_name TEXT NOT NULL,
    realm_name TEXT NOT NULL,
    race TEXT NOT NULL,
    class TEXT NOT NULL,
    spec TEXT NOT NULL,
    gender TEXT NOT NULL,
    faction TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(realm_name, character_name)
);
