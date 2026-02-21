CREATE TABLE poems(
    slug VARCHAR PRIMARY KEY,
    title VARCHAR NOT NULL,
    poem TEXT,
    notes TEXT,
    published TIMESTAMPTZ DEFAULT NOW(),
    modified TIMESTAMPTZ DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION poems_insert_slug() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO slugs (slug, content_type)
    VALUES (NEW.slug, 'poem');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION poems_update_slug() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.slug != NEW.slug THEN
        DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'poem';
        INSERT INTO slugs (slug, content_type)
        VALUES (NEW.slug, 'poem');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION poems_delete_slug() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'poem';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER poems_after_insert
    AFTER INSERT ON poems
    FOR EACH ROW
    EXECUTE FUNCTION poems_insert_slug();

CREATE TRIGGER poems_after_update
    AFTER UPDATE ON poems
    FOR EACH ROW
    EXECUTE FUNCTION poems_update_slug();

CREATE TRIGGER poems_after_delete
    AFTER DELETE ON poems
    FOR EACH ROW
    EXECUTE FUNCTION poems_delete_slug();