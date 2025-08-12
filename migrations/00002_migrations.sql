-- Combined Migrations Version: 00002
-- This file contains all migrations for this version

-- Migration for table: tests
CREATE TABLE IF NOT EXISTS tests (
	ID INTEGER PRIMARY KEY,
	Color VARCHAR,
	Stock INTEGER,
	Cents INTEGER,
	Price DOUBLE PRECISION,
	ImagePath VARCHAR,
	PRIMARY KEY (ID)
);
