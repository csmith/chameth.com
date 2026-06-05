ALTER TABLE syndications
    ADD COLUMN disposition VARCHAR NOT NULL DEFAULT 'anchor',
    ADD COLUMN rel VARCHAR;
