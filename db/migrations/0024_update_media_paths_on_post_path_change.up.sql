-- Trigger function to update media relations when post path changes
CREATE OR REPLACE FUNCTION posts_update_media_paths() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.path != NEW.path THEN
        -- Update all media relations for this post using string replacement
        UPDATE media_relations
        SET path = REPLACE(path, OLD.path, NEW.path)
        WHERE entity_type = 'post' AND entity_id = NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to fire after update on posts
DROP TRIGGER IF EXISTS posts_update_media_paths_trigger ON posts;

CREATE TRIGGER posts_update_media_paths_trigger
    AFTER UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION posts_update_media_paths();
