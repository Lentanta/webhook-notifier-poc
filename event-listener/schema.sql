-- Events table
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    event_name VARCHAR(100) NOT NULL,
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    payload JSONB NOT NULL,
    webhook_id VARCHAR(100) NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient querying of unprocessed events
CREATE INDEX IF NOT EXISTS idx_events_unprocessed
ON events (processed, created_at)
WHERE processed = false;

-- Trigger function to notify on new events
CREATE OR REPLACE FUNCTION notify_new_event()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('new_event', NEW.id::TEXT);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to send notification on insert
DROP TRIGGER IF EXISTS event_insert_trigger ON events;
CREATE TRIGGER event_insert_trigger
    AFTER INSERT ON events
    FOR EACH ROW
    EXECUTE FUNCTION notify_new_event();
