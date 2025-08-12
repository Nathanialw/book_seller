-- Combined Migrations Version: 00001
-- This file contains all migrations for this version

-- Undo migration for table: cart_items
DROP TABLE IF EXISTS cart_items CASCADE;

-- Undo migration for table: cart_itemses
DROP TABLE IF EXISTS cart_itemses CASCADE;

-- Undo migration for table: carts
DROP TABLE IF EXISTS carts CASCADE;

-- Undo migration for table: orders
DROP TABLE IF EXISTS orders CASCADE;

-- Undo migration for table: order_items
DROP TABLE IF EXISTS order_items CASCADE;

-- Undo migration for table: products
DROP TABLE IF EXISTS products CASCADE;

-- Undo migration for table: variants
DROP TABLE IF EXISTS variants CASCADE;
