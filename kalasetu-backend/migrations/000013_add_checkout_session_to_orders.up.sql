-- Links an Order back to the Checkout Session it was fulfilled from.
-- checkout_session_id is unique so fulfilment stays idempotent: a session can
-- produce at most one Order. charge_id (existing column) now holds the
-- gateway's payment identifier instead of the old PaymentProvider's charge id.
ALTER TABLE orders ADD COLUMN checkout_session_id INTEGER REFERENCES checkout_sessions(id);

CREATE UNIQUE INDEX orders_checkout_session_id_idx
    ON orders (checkout_session_id) WHERE checkout_session_id IS NOT NULL;
