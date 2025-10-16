CREATE TABLE posts (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    slug VARCHAR NOT NULL UNIQUE,
    title VARCHAR NOT NULL,
    content TEXT NOT NULL,
    date DATE NOT NULL,
    format VARCHAR NOT NULL,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX idx_posts_slug ON posts(slug);
CREATE INDEX idx_posts_date ON posts(date DESC);

CREATE OR REPLACE FUNCTION posts_insert_slug() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO slugs (slug, content_type)
    VALUES (NEW.slug, 'post');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION posts_update_slug() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.slug != NEW.slug THEN
        DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'post';
        INSERT INTO slugs (slug, content_type)
        VALUES (NEW.slug, 'post');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION posts_delete_slug() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'post';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER posts_after_insert
    AFTER INSERT ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_insert_slug();

CREATE TRIGGER posts_after_update
    AFTER UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_update_slug();

CREATE TRIGGER posts_after_delete
    AFTER DELETE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_delete_slug();
