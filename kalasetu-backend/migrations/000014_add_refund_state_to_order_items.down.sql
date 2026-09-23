ALTER TABLE order_items
    DROP COLUMN IF EXISTS refund_status,
    DROP COLUMN IF EXISTS refund_id;
