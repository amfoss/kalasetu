DROP TABLE IF EXISTS listing_images;
DROP TABLE IF EXISTS listings;

DELETE FROM categories WHERE category_name IN
    ('Pottery', 'Textiles', 'Woodwork', 'Jewellery', 'Paintings', 'Metalwork');
