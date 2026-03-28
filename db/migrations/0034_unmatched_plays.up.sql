CREATE TABLE unmatched_music_plays (
    id               INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    play_id          TEXT NOT NULL UNIQUE,
    music_brainz_id  TEXT NOT NULL,
    title            TEXT NOT NULL,
    played_at        TIMESTAMP NOT NULL
);
