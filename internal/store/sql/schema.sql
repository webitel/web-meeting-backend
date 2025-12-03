CREATE SCHEMA IF NOT EXISTS meetings;

CREATE TABLE IF NOT EXISTS meetings.web_meetings (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL,
    variables JSONB,
    url TEXT
);

create index web_meetings_expires_at_index
    on meetings.web_meetings (expires_at);