-- A late payment whose Checkout Session could not be re-fulfilled (its
-- reserved Stock is genuinely gone) is refunded automatically rather than
-- left with KalaSetu; the refund is recorded on the session itself, the way
-- a cancelled Order Item's refund is recorded separately from its status.
ALTER TABLE checkout_sessions
    ADD COLUMN refund_id VARCHAR(255),
    ADD COLUMN refund_status VARCHAR(20)
        CHECK (refund_status IS NULL OR refund_status IN ('accepted', 'settled', 'failed')),
    ADD COLUMN refund_reason VARCHAR(255);

-- A payment the gateway reports for a gateway order id no Checkout Session
-- was ever opened for is recorded here and refunded, rather than silently
-- ignored.
CREATE TABLE unmatched_payments (
    id SERIAL PRIMARY KEY,
    gateway_order_id VARCHAR(255) NOT NULL,
    payment_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    refund_id VARCHAR(255),
    refund_status VARCHAR(20)
        CHECK (refund_status IS NULL OR refund_status IN ('accepted', 'settled', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
