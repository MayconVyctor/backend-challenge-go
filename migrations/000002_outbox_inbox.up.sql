CREATE TABLE IF NOT EXISTS inbox (
    id UUID PRIMARY KEY,
    consumer_name VARCHAR(255) NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_inbox_message UNIQUE (consumer_name, message_id)
);

CREATE TABLE IF NOT EXISTS outbox (
    event_id UUID PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL,
    aggregate_id UUID NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE
);
