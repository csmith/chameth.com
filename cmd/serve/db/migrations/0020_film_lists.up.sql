CREATE TABLE film_lists(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR NOT NULL,
    description TEXT,
    published BOOLEAN DEFAULT false,
    path VARCHAR UNIQUE
);

CREATE INDEX idx_film_lists_path ON film_lists(path);

CREATE TABLE film_list_entries(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    film_list_id INTEGER NOT NULL REFERENCES film_lists(id) ON DELETE CASCADE,
    film_id INTEGER NOT NULL REFERENCES films(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    UNIQUE (film_list_id, position) DEFERRABLE INITIALLY IMMEDIATE
);

CREATE INDEX idx_film_list_entries_list_id ON film_list_entries(film_list_id);
CREATE INDEX idx_film_list_entries_film_id ON film_list_entries(film_id);

CREATE OR REPLACE FUNCTION film_lists_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'film_list');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION film_lists_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'film_list';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'film_list');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION film_lists_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'film_list';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER film_lists_after_insert
    AFTER INSERT ON film_lists
    FOR EACH ROW
    EXECUTE FUNCTION film_lists_insert_path();

CREATE TRIGGER film_lists_after_update
    AFTER UPDATE OF path ON film_lists
    FOR EACH ROW
    EXECUTE FUNCTION film_lists_update_path();

CREATE TRIGGER film_lists_after_delete
    AFTER DELETE ON film_lists
    FOR EACH ROW
    EXECUTE FUNCTION film_lists_delete_path();
