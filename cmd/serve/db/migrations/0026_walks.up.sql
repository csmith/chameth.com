CREATE TABLE walks (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    external_id UUID NOT NULL UNIQUE,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_seconds DOUBLE PRECISION NOT NULL,
    distance_km DOUBLE PRECISION NOT NULL,
    elevation_gain_meters DOUBLE PRECISION NOT NULL
);

CREATE INDEX idx_walks_external_id ON walks(external_id);
CREATE INDEX idx_walks_start_date ON walks(start_date DESC);
