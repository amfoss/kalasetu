INSERT INTO categories (category_name) VALUES
    ('Pottery'),
    ('Textiles'),
    ('Woodwork'),
    ('Jewellery'),
    ('Paintings'),
    ('Metalwork')
ON CONFLICT (category_name) DO NOTHING;

-- Price is exact 2-decimal INR; stock of 1 is a one-of-a-kind piece.
CREATE TABLE listings (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    title VARCHAR(255) NOT NULL CHECK (btrim(title) <> ''),
    description TEXT NOT NULL DEFAULT '',
    price NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    currency CHAR(3) NOT NULL DEFAULT 'INR' CHECK (currency = 'INR'),
    stock INTEGER NOT NULL CHECK (stock >= 0),
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX listings_seller_id_idx ON listings (seller_id);
CREATE INDEX listings_category_id_idx ON listings (category_id);

-- position 0 is the cover image.
CREATE TABLE listing_images (
    listing_id INTEGER NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position BETWEEN 0 AND 7),
    url VARCHAR(2048) NOT NULL,
    PRIMARY KEY (listing_id, position)
);
