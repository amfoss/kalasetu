DROP INDEX IF EXISTS orders_checkout_session_id_idx;
ALTER TABLE orders DROP COLUMN IF EXISTS checkout_session_id;
