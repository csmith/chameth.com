CREATE TABLE music_plays (
    id        INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    play_id   TEXT NOT NULL UNIQUE,
    track_id  INTEGER NOT NULL REFERENCES music_tracks(id),
    played_at TIMESTAMP NOT NULL
);
