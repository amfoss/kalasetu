-- An Order is one checkout of a Cart. The shipping address is a snapshot; the
-- Order has no status of its own, each Order Item carries a fulfilment status.
CREATE TABLE orders (
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
    charge_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX orders_buyer_id_idx ON orders (buyer_id);

-- title and price are immutable snapshots of the Listing at purchase time.
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    listing_id INTEGER NOT NULL REFERENCES listings(id),
    seller_id INTEGER NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    price NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    status VARCHAR(20) NOT NULL
        CHECK (status IN ('paid', 'shipped', 'delivered', 'cancelled'))
);

CREATE INDEX order_items_order_id_idx ON order_items (order_id);
CREATE INDEX order_items_seller_id_idx ON order_items (seller_id);
