-- Rename slugs table to paths
ALTER TABLE slugs RENAME TO paths;
ALTER TABLE paths RENAME COLUMN slug TO path;

-- Rename slug columns in content tables
ALTER TABLE poems RENAME COLUMN slug TO path;
ALTER TABLE snippets RENAME COLUMN slug TO path;
ALTER TABLE staticpages RENAME COLUMN slug TO path;
ALTER TABLE posts RENAME COLUMN slug TO path;
ALTER TABLE media_relations RENAME COLUMN slug TO path;

-- Rename indexes
ALTER INDEX idx_posts_slug RENAME TO idx_posts_path;

-- Rename constraints
ALTER TABLE snippets RENAME CONSTRAINT snippets_slug_unique TO snippets_path_unique;
ALTER TABLE poems RENAME CONSTRAINT poems_slug_unique TO poems_path_unique;

-- Update trigger functions for poems
CREATE OR REPLACE FUNCTION poems_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'poem');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION poems_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'poem';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'poem');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION poems_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'poem';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Update trigger functions for snippets
CREATE OR REPLACE FUNCTION snippets_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'snippet');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION snippets_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'snippet';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'snippet');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION snippets_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'snippet';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Update trigger functions for staticpages
CREATE OR REPLACE FUNCTION staticpages_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'staticpage');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION staticpages_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'staticpage';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'staticpage');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION staticpages_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'staticpage';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Update trigger functions for posts
CREATE OR REPLACE FUNCTION posts_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'post');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION posts_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'post';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'post');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION posts_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'post';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Update trigger functions for media_relations
CREATE OR REPLACE FUNCTION media_relations_insert_path() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO paths (path, content_type)
    VALUES (NEW.path, 'media');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION media_relations_update_path() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        DELETE FROM paths WHERE path = OLD.path AND content_type = 'media';
        INSERT INTO paths (path, content_type)
        VALUES (NEW.path, 'media');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION media_relations_delete_path() RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM paths WHERE path = OLD.path AND content_type = 'media';
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Drop old triggers and create new ones for poems
DROP TRIGGER poems_after_insert ON poems;
DROP TRIGGER poems_after_update ON poems;
DROP TRIGGER poems_after_delete ON poems;

CREATE TRIGGER poems_after_insert
    AFTER INSERT ON poems
    FOR EACH ROW
    EXECUTE FUNCTION poems_insert_path();

CREATE TRIGGER poems_after_update
    AFTER UPDATE ON poems
    FOR EACH ROW
    EXECUTE FUNCTION poems_update_path();

CREATE TRIGGER poems_after_delete
    AFTER DELETE ON poems
    FOR EACH ROW
    EXECUTE FUNCTION poems_delete_path();

-- Drop old triggers and create new ones for snippets
DROP TRIGGER snippets_after_insert ON snippets;
DROP TRIGGER snippets_after_update ON snippets;
DROP TRIGGER snippets_after_delete ON snippets;

CREATE TRIGGER snippets_after_insert
    AFTER INSERT ON snippets
    FOR EACH ROW
    EXECUTE FUNCTION snippets_insert_path();

CREATE TRIGGER snippets_after_update
    AFTER UPDATE ON snippets
    FOR EACH ROW
    EXECUTE FUNCTION snippets_update_path();

CREATE TRIGGER snippets_after_delete
    AFTER DELETE ON snippets
    FOR EACH ROW
    EXECUTE FUNCTION snippets_delete_path();

-- Drop old triggers and create new ones for staticpages
DROP TRIGGER staticpages_after_insert ON staticpages;
DROP TRIGGER staticpages_after_update ON staticpages;
DROP TRIGGER staticpages_after_delete ON staticpages;

CREATE TRIGGER staticpages_after_insert
    AFTER INSERT ON staticpages
    FOR EACH ROW
    EXECUTE FUNCTION staticpages_insert_path();

CREATE TRIGGER staticpages_after_update
    AFTER UPDATE ON staticpages
    FOR EACH ROW
    EXECUTE FUNCTION staticpages_update_path();

CREATE TRIGGER staticpages_after_delete
    AFTER DELETE ON staticpages
    FOR EACH ROW
    EXECUTE FUNCTION staticpages_delete_path();

-- Drop old triggers and create new ones for posts
DROP TRIGGER posts_after_insert ON posts;
DROP TRIGGER posts_after_update ON posts;
DROP TRIGGER posts_after_delete ON posts;

CREATE TRIGGER posts_after_insert
    AFTER INSERT ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_insert_path();

CREATE TRIGGER posts_after_update
    AFTER UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_update_path();

CREATE TRIGGER posts_after_delete
    AFTER DELETE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_delete_path();

-- Drop old triggers and create new ones for media_relations
DROP TRIGGER media_relations_after_insert ON media_relations;
DROP TRIGGER media_relations_after_update ON media_relations;
DROP TRIGGER media_relations_after_delete ON media_relations;

CREATE TRIGGER media_relations_after_insert
    AFTER INSERT ON media_relations
    FOR EACH ROW
    EXECUTE FUNCTION media_relations_insert_path();

CREATE TRIGGER media_relations_after_update
    AFTER UPDATE ON media_relations
    FOR EACH ROW
    EXECUTE FUNCTION media_relations_update_path();

CREATE TRIGGER media_relations_after_delete
    AFTER DELETE ON media_relations
    FOR EACH ROW
    EXECUTE FUNCTION media_relations_delete_path();

-- Drop old trigger functions
DROP FUNCTION poems_insert_slug();
DROP FUNCTION poems_update_slug();
DROP FUNCTION poems_delete_slug();
DROP FUNCTION snippets_insert_slug();
DROP FUNCTION snippets_update_slug();
DROP FUNCTION snippets_delete_slug();
DROP FUNCTION staticpages_insert_slug();
DROP FUNCTION staticpages_update_slug();
DROP FUNCTION staticpages_delete_slug();
DROP FUNCTION posts_insert_slug();
DROP FUNCTION posts_update_slug();
DROP FUNCTION posts_delete_slug();
DROP FUNCTION media_relations_insert_slug();
DROP FUNCTION media_relations_update_slug();
DROP FUNCTION media_relations_delete_slug();
