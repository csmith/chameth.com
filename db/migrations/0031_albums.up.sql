CREATE TABLE music_albums (
    id              INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    music_brainz_id TEXT NOT NULL UNIQUE,
    subsonic_id     TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    sort_name       TEXT NOT NULL,
    year            INTEGER,
    artist_id       INTEGER REFERENCES music_artists(id)
);
