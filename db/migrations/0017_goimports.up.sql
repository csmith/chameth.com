ALTER TABLE paths ADD COLUMN prefix_match BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE goimports (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    path VARCHAR NOT NULL UNIQUE,
    vcs VARCHAR NOT NULL,
    repo_url VARCHAR NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_goimports_path ON goimports(path);

CREATE OR REPLACE FUNCTION goimports_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type, prefix_match)
    VALUES (NEW.path, 'goimport', TRUE);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION goimports_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'goimport';
        INSERT INTO paths (path, content_type, prefix_match)
        VALUES (NEW.path, 'goimport', TRUE);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION goimports_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'goimport';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER goimports_after_insert
    AFTER INSERT ON goimports
    FOR EACH ROW
    EXECUTE FUNCTION goimports_insert_path();

CREATE TRIGGER goimports_after_update
    AFTER UPDATE ON goimports
    FOR EACH ROW
    EXECUTE FUNCTION goimports_update_path();

CREATE TRIGGER goimports_after_delete
    AFTER DELETE ON goimports
    FOR EACH ROW
    EXECUTE FUNCTION goimports_delete_path();
