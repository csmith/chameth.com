CREATE TABLE pastes (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    path VARCHAR NOT NULL UNIQUE,
    title VARCHAR NOT NULL,
    language VARCHAR,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    date TIMESTAMPTZ NOT NULL,
    content TEXT NOT NULL
);

CREATE INDEX idx_pastes_path ON pastes(path);
CREATE INDEX idx_pastes_date ON pastes(date DESC);

-- Trigger function for INSERT
CREATE OR REPLACE FUNCTION pastes_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'paste');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger function for UPDATE
CREATE OR REPLACE FUNCTION pastes_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'paste';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'paste');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger function for DELETE
CREATE OR REPLACE FUNCTION pastes_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'paste';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
CREATE TRIGGER pastes_after_insert
    AFTER INSERT ON pastes
    FOR EACH ROW
EXECUTE FUNCTION pastes_insert_path();

CREATE TRIGGER pastes_after_update
    AFTER UPDATE ON pastes
    FOR EACH ROW
EXECUTE FUNCTION pastes_update_path();

CREATE TRIGGER pastes_after_delete
    AFTER DELETE ON pastes
    FOR EACH ROW
EXECUTE FUNCTION pastes_delete_path();
