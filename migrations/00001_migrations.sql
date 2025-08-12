-- Combined Migrations Version: 00001
-- This file contains all migrations for this version

-- Migration for table: cart_items
CREATE TABLE IF NOT EXISTS cart_items (
	VariantID INTEGER,
	Quantity INTEGER
);

-- Migration for table: cart_itemses
CREATE TABLE IF NOT EXISTS cart_itemses (
	Variant TEXT,
	Name VARCHAR,
	Quantity INTEGER,
	Total DOUBLE PRECISION
);

-- Migration for table: carts
CREATE TABLE IF NOT EXISTS carts (
	Products TEXT,
	Subtotal DOUBLE PRECISION,
	Tax DOUBLE PRECISION,
	Total DOUBLE PRECISION
);

-- Migration for table: orders
CREATE TABLE IF NOT EXISTS orders (
	ID INTEGER PRIMARY KEY,
	OrderNumber VARCHAR,
	Email VARCHAR,
	Address VARCHAR,
	City VARCHAR,
	PostalCode VARCHAR,
	Country VARCHAR,
	Products TEXT,
	PRIMARY KEY (ID)
);

-- Migration for table: order_items
CREATE TABLE IF NOT EXISTS order_items (
	VariantID INTEGER,
	Quantity INTEGER,
	Cents INTEGER,
	Price DOUBLE PRECISION,
	ProductTitle VARCHAR,
	ProductTitles VARCHAR,
	ProductTitlessasd VARCHAR,
	VariantColor VARCHAR
);

-- Migration for table: products
CREATE TABLE IF NOT EXISTS products (
	ID INTEGER PRIMARY KEY,
	Title VARCHAR,
	Author VARCHAR,
	Description VARCHAR,
	LowestPrice DOUBLE PRECISION,
	Type0 VARCHAR,
	Variants TEXT,
	PRIMARY KEY (ID)
);

-- Migration for table: variants
CREATE TABLE IF NOT EXISTS variants (
	ID INTEGER PRIMARY KEY,
	Color VARCHAR,
	Stock INTEGER,
	Cents INTEGER,
	Price DOUBLE PRECISION,
	ImagePath VARCHAR,
	PRIMARY KEY (ID)
);
