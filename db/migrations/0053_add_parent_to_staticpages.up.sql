ALTER TABLE staticpages
    ADD COLUMN parent_id INTEGER REFERENCES staticpages(id) ON DELETE SET NULL;
