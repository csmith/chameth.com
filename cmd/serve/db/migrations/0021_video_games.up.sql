CREATE TABLE video_games (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR NOT NULL,
    platform VARCHAR,
    overview TEXT,
    published BOOLEAN DEFAULT false,
    path VARCHAR UNIQUE
);

CREATE INDEX idx_video_games_path ON video_games(path);

CREATE TABLE video_game_reviews (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    video_game_id INTEGER NOT NULL REFERENCES video_games(id) ON DELETE CASCADE,
    played_date DATE NOT NULL DEFAULT CURRENT_DATE,
    rating INTEGER NOT NULL,
    playtime INTEGER,
    completion_status VARCHAR,
    notes TEXT,
    published BOOLEAN DEFAULT false
);

CREATE INDEX idx_video_game_reviews_video_game_id ON video_game_reviews(video_game_id);

-- Create trigger functions for video_games
CREATE OR REPLACE FUNCTION video_games_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'videogame');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION video_games_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'videogame';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'videogame');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION video_games_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'videogame';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for video_games
CREATE TRIGGER video_games_after_insert
    AFTER INSERT ON video_games
    FOR EACH ROW
    EXECUTE FUNCTION video_games_insert_path();

CREATE TRIGGER video_games_after_update
    AFTER UPDATE ON video_games
    FOR EACH ROW
    EXECUTE FUNCTION video_games_update_path();

CREATE TRIGGER video_games_after_delete
    AFTER DELETE ON video_games
    FOR EACH ROW
    EXECUTE FUNCTION video_games_delete_path();
