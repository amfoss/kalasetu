-- Every gateway webhook delivery is recorded by its identifier before it is
-- acted on, so a redelivery (Razorpay retries deliveries) is recognised and
-- turned into a no-op rather than repeating whatever the first delivery did.
CREATE TABLE webhook_events (
    event_id VARCHAR(64) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
