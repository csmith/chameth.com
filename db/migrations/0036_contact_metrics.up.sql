CREATE TABLE contact_metrics (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    method TEXT NOT NULL,
    user_agent TEXT,
    checks JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    page TEXT,
    sender_name TEXT,
    sender_email TEXT,
    message TEXT
);
