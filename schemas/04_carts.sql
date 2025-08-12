-- Migration for table: cart_items
CREATE TABLE IF NOT EXISTS cart_items (
	id SERIAL PRIMARY KEY,
	variant_id INTEGER,
	name VARCHAR,
	quantity INTEGER,
	total DOUBLE PRECISION,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT fk_cart_items_variant_id_variants FOREIGN KEY (variant_id) REFERENCES variants(ID)
);


-- Migration for table: carts
CREATE TABLE IF NOT EXISTS carts (
	id SERIAL PRIMARY KEY,
	subtotal DOUBLE PRECISION,
	tax DOUBLE PRECISION,
	total DOUBLE PRECISION,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
