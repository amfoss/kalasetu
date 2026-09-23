-- Cancelling an Order Item now records the gateway's refund identifier and
-- the refund's own state, separate from the item's fulfilment status: a
-- cancelled item's refund can be accepted, then later settled or failed
-- without the fulfilment status changing.
ALTER TABLE order_items
    ADD COLUMN refund_id VARCHAR(255),
    ADD COLUMN refund_status VARCHAR(20)
        CHECK (refund_status IS NULL OR refund_status IN ('accepted', 'settled', 'failed'));
