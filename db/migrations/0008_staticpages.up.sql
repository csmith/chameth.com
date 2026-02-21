CREATE TABLE staticpages(
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    slug VARCHAR NOT NULL UNIQUE,
    title VARCHAR NOT NULL,
    content TEXT NOT NULL
);

CREATE OR REPLACE FUNCTION staticpages_insert_slug() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO slugs (slug, content_type)
    VALUES (NEW.slug, 'staticpage');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION staticpages_update_slug() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.slug != NEW.slug THEN
        DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'staticpage';
        INSERT INTO slugs (slug, content_type)
        VALUES (NEW.slug, 'staticpage');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION staticpages_delete_slug() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'staticpage';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER staticpages_after_insert
    AFTER INSERT ON staticpages
    FOR EACH ROW
    EXECUTE FUNCTION staticpages_insert_slug();

CREATE TRIGGER staticpages_after_update
    AFTER UPDATE ON staticpages
    FOR EACH ROW
    EXECUTE FUNCTION staticpages_update_slug();

CREATE TRIGGER staticpages_after_delete
    AFTER DELETE ON staticpages
    FOR EACH ROW
    EXECUTE FUNCTION staticpages_delete_slug();
