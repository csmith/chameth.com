CREATE TABLE snippets(
    slug VARCHAR PRIMARY KEY,
    title VARCHAR NOT NULL,
    topic VARCHAR,
    content TEXT
);

CREATE OR REPLACE FUNCTION snippets_insert_slug() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO slugs (slug, content_type)
    VALUES (NEW.slug, 'snippet');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION snippets_update_slug() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.slug != NEW.slug THEN
        DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'snippet';
        INSERT INTO slugs (slug, content_type)
        VALUES (NEW.slug, 'snippet');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION snippets_delete_slug() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'snippet';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER snippets_after_insert
    AFTER INSERT ON snippets
    FOR EACH ROW
    EXECUTE FUNCTION snippets_insert_slug();

CREATE TRIGGER snippets_after_update
    AFTER UPDATE ON snippets
    FOR EACH ROW
    EXECUTE FUNCTION snippets_update_slug();

CREATE TRIGGER snippets_after_delete
    AFTER DELETE ON snippets
    FOR EACH ROW
    EXECUTE FUNCTION snippets_delete_slug();
