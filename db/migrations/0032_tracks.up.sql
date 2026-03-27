CREATE TABLE music_tracks (
    id              INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    subsonic_id     TEXT NOT NULL UNIQUE,
    music_brainz_id TEXT NOT NULL,
    album_id        INTEGER NOT NULL REFERENCES music_albums(id),
    name            TEXT NOT NULL,
    duration        INTEGER,
    disc_number     INTEGER,
    track_number    INTEGER
);
