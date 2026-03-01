ALTER TABLE staticpages
    ADD COLUMN sitemap_frequency VARCHAR DEFAULT NULL,
    ADD COLUMN sitemap_priority NUMERIC(2,1) DEFAULT NULL;
