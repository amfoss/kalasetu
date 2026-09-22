-- A Checkout Session is a Buyer's reserved, time-limited intent to purchase:
-- a snapshot of the Cart lines and Shipping address, the total, the gateway
-- order id, and an expiry. Stock is decremented when the session is created,
-- not when it is paid. Only one open session may exist per Buyer at a time.
CREATE TABLE checkout_sessions (
    id SERIAL PRIMARY KEY,
    buyer_id INTEGER NOT NULL REFERENCES users(id),
    ship_name VARCHAR(255) NOT NULL,
    ship_phone VARCHAR(50) NOT NULL,
    ship_line1 VARCHAR(255) NOT NULL,
    ship_line2 VARCHAR(255) NOT NULL DEFAULT '',
    ship_city VARCHAR(255) NOT NULL,
    ship_state VARCHAR(255) NOT NULL,
    ship_postal_code VARCHAR(20) NOT NULL,
    ship_country VARCHAR(100) NOT NULL,
    total NUMERIC(12, 2) NOT NULL CHECK (total > 0),
    gateway_order_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'consumed', 'cancelled', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX checkout_sessions_buyer_id_idx ON checkout_sessions (buyer_id);

-- A Buyer with a live Checkout Session who starts another is rejected rather
-- than having the first superseded; this is the database-level backstop for
-- that rule.
CREATE UNIQUE INDEX checkout_sessions_one_open_per_buyer
    ON checkout_sessions (buyer_id) WHERE status = 'open';

-- title, price and quantity are immutable snapshots of the Listing at the
-- moment the Checkout Session was created.
CREATE TABLE checkout_session_items (
    id SERIAL PRIMARY KEY,
    checkout_session_id INTEGER NOT NULL REFERENCES checkout_sessions(id) ON DELETE CASCADE,
    listing_id INTEGER NOT NULL REFERENCES listings(id),
    seller_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    price NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    quantity INTEGER NOT NULL CHECK (quantity > 0)
);

CREATE INDEX checkout_session_items_session_id_idx ON checkout_session_items (checkout_session_id);
