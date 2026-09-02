CREATE TABLE shortcode_data (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shortcode TEXT NOT NULL,
    version INTEGER NOT NULL,
    args_hash TEXT NOT NULL,
    args JSONB NOT NULL,
    data JSONB NOT NULL,
    next_refresh_at TIMESTAMP WITH TIME ZONE,
    retrieved_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (shortcode, version, args_hash)
);

CREATE INDEX idx_shortcode_data_due ON shortcode_data(next_refresh_at)
    WHERE next_refresh_at IS NOT NULL;
