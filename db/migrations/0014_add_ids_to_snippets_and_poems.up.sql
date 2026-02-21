-- Add ID columns to snippets and poems tables and make them the primary key

-- Snippets table
ALTER TABLE snippets ADD COLUMN id INTEGER GENERATED ALWAYS AS IDENTITY;
ALTER TABLE snippets DROP CONSTRAINT snippets_pkey;
ALTER TABLE snippets ADD PRIMARY KEY (id);
ALTER TABLE snippets ADD CONSTRAINT snippets_slug_unique UNIQUE (slug);

-- Poems table
ALTER TABLE poems ADD COLUMN id INTEGER GENERATED ALWAYS AS IDENTITY;
ALTER TABLE poems DROP CONSTRAINT poems_pkey;
ALTER TABLE poems ADD PRIMARY KEY (id);
ALTER TABLE poems ADD CONSTRAINT poems_slug_unique UNIQUE (slug);
