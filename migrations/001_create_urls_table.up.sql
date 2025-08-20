CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    code TEXT UNIQUE,
    long_url TEXT NOT NULL,
    expire_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);