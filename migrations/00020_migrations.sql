-- Migration for table: cart_items
CREATE TABLE IF NOT EXISTS cart_items (
	ID INTEGER PRIMARY KEY,
	VariantID INTEGER,
	Quantity INTEGER
);

-- Migration for table: cart_itemses
CREATE TABLE IF NOT EXISTS cart_itemses (
	ID INTEGER PRIMARY KEY,
	Variant TEXT,
	Name VARCHAR,
	Quantity INTEGER,
	Total DOUBLE PRECISION
);

-- Migration for table: carts
CREATE TABLE IF NOT EXISTS carts (
	ID INTEGER PRIMARY KEY,
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
	Products TEXT
);

-- Migration for table: order_items
CREATE TABLE IF NOT EXISTS order_items (
	ID INTEGER PRIMARY KEY,
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
	Variants TEXT
);

-- Migration for table: variants
CREATE TABLE IF NOT EXISTS variants (
	ID INTEGER PRIMARY KEY,
	ProductID INTEGER,
	Color VARCHAR,
	Stock INTEGER,
	Cents INTEGER,
	Price DOUBLE PRECISION,
	ImagePath VARCHAR,
	CONSTRAINT fk_variants_ProductID_products FOREIGN KEY (ProductID) REFERENCES products(ID)
);