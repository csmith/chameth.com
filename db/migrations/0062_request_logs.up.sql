CREATE TABLE request_logs (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    url TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    ip_hash BYTEA,
    asn BIGINT,
    as_org TEXT,
    duration_us BIGINT NOT NULL,
    response_bytes BIGINT NOT NULL,
    status INTEGER NOT NULL
);
