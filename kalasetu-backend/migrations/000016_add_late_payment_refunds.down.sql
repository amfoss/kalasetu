DROP TABLE unmatched_payments;

ALTER TABLE checkout_sessions
    DROP COLUMN refund_id,
    DROP COLUMN refund_status,
    DROP COLUMN refund_reason;
