CREATE TABLE boardgame_games (
    id UUID PRIMARY KEY,
    bgg_id INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL,
    year INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'unowned' CHECK (status IN ('unowned', 'owned', 'sold'))
);

CREATE TABLE boardgame_plays (
    id UUID PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES boardgame_games(id),
    date TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_boardgame_plays_game_id ON boardgame_plays(game_id);
CREATE INDEX idx_boardgame_plays_date ON boardgame_plays(date DESC);
