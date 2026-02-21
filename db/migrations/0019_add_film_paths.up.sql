-- Add path column to films table (without unique constraint initially)
ALTER TABLE films ADD COLUMN path VARCHAR;

-- Update existing films with paths based on title and year
UPDATE films
SET path = '/films/' || 
    regexp_replace(
        regexp_replace(
            regexp_replace(
                lower(title),
                '[^a-z0-9]+', '-', 'g'
            ),
            '^-+|-+$', '', 'g'
        ),
        '-+', '-', 'g'
    ) || '-' || year || '/'
WHERE path IS NULL;

-- Add unique constraint to path column
ALTER TABLE films ADD CONSTRAINT films_path_unique UNIQUE (path);

-- Create index on path
CREATE INDEX idx_films_path ON films(path);

-- Insert all existing film paths into the paths table
INSERT INTO paths (path, content_type)
SELECT path, 'film'
FROM films
ON CONFLICT (path) DO NOTHING;

-- Create trigger functions for films
CREATE OR REPLACE FUNCTION films_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'film');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION films_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'film';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'film');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION films_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'film';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for films
CREATE TRIGGER films_after_insert
    AFTER INSERT ON films
    FOR EACH ROW
    EXECUTE FUNCTION films_insert_path();

CREATE TRIGGER films_after_update
    AFTER UPDATE ON films
    FOR EACH ROW
    EXECUTE FUNCTION films_update_path();

CREATE TRIGGER films_after_delete
    AFTER DELETE ON films
    FOR EACH ROW
    EXECUTE FUNCTION films_delete_path();
