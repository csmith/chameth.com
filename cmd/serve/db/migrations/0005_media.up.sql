CREATE TABLE media
(
    id                INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    content_type      VARCHAR NOT NULL,
    original_filename VARCHAR NOT NULL,
    data              BYTEA   NOT NULL
);

CREATE TABLE media_relations
(
    slug        VARCHAR PRIMARY KEY,
    media_id    INTEGER NOT NULL REFERENCES media (id) ON DELETE CASCADE,
    description TEXT,
    caption     TEXT,
    role        VARCHAR,
    entity_type VARCHAR NOT NULL,
    entity_id   INT NOT NULL
);

CREATE OR REPLACE FUNCTION media_relations_insert_slug() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO slugs (slug, content_type)
    VALUES (NEW.slug, 'media');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION media_relations_update_slug() RETURNS TRIGGER AS
$$
BEGIN
    IF OLD.slug != NEW.slug THEN
        DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'media';
        INSERT INTO slugs (slug, content_type)
        VALUES (NEW.slug, 'media');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION media_relations_delete_slug() RETURNS TRIGGER AS
$$
BEGIN
    DELETE FROM slugs WHERE slug = OLD.slug AND content_type = 'media';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER media_relations_after_insert
    AFTER INSERT
    ON media_relations
    FOR EACH ROW
EXECUTE FUNCTION media_relations_insert_slug();

CREATE TRIGGER media_relations_after_update
    AFTER UPDATE
    ON media_relations
    FOR EACH ROW
EXECUTE FUNCTION media_relations_update_slug();

CREATE TRIGGER media_relations_after_delete
    AFTER DELETE
    ON media_relations
    FOR EACH ROW
EXECUTE FUNCTION media_relations_delete_slug();
