CREATE TABLE music_artists (
    id             INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    music_brainz_id TEXT NOT NULL UNIQUE,
    subsonic_id    TEXT NOT NULL UNIQUE,
    name           TEXT NOT NULL,
    sort_name      TEXT NOT NULL
);
