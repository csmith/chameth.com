-- Add published field to all content tables
-- First rename the existing timestamp column in poems to date and drop modified
ALTER TABLE poems RENAME COLUMN published TO date;
ALTER TABLE poems DROP COLUMN modified;
ALTER TABLE poems ADD COLUMN published BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE posts ADD COLUMN published BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE prints ADD COLUMN published BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE projects ADD COLUMN published BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE snippets ADD COLUMN published BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE staticpages ADD COLUMN published BOOLEAN NOT NULL DEFAULT FALSE;

-- Update all existing rows to be published
UPDATE poems SET published = TRUE;
UPDATE posts SET published = TRUE;
UPDATE prints SET published = TRUE;
UPDATE projects SET published = TRUE;
UPDATE snippets SET published = TRUE;
UPDATE staticpages SET published = TRUE;
