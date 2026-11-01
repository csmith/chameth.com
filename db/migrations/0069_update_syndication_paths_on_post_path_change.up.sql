CREATE OR REPLACE FUNCTION posts_update_syndication_paths() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path IS DISTINCT FROM NEW.path THEN
        UPDATE syndications SET path = NEW.path WHERE path = OLD.path;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER posts_update_syndication_paths_trigger
    AFTER UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_update_syndication_paths();
