#!/bin/bash

# --- CONFIG ---
DB_NAME="ecommerce"
DB_USER="admin"
DB_PASS="securepassword"  # ⚠️ Change this in production!

# Install PostgreSQL if not installed
if ! command -v psql &> /dev/null; then
    echo "Installing PostgreSQL..."
    sudo apt update
    sudo apt install -y postgresql postgresql-contrib
fi

# Start and enable PostgreSQL service
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Function: Check if database exists
db_exists() {
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='$1'" | grep -q 1
}

# Function: Check if user exists
user_exists() {
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='$1'" | grep -q 1
}

# Create database if not exists
if db_exists "$DB_NAME"; then
  echo "✅ Database '$DB_NAME' already exists."
else
  echo "🔧 Creating database '$DB_NAME'..."
  sudo -u postgres createdb "$DB_NAME"
fi

# Create user if not exists
if user_exists "$DB_USER"; then
  echo "✅ User '$DB_USER' already exists."
else
  echo "🔧 Creating user '$DB_USER'..."
  sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';"
fi

# Grant privileges to user on database
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# Create tables and extensions
echo "📦 Setting up schema..."
sudo -u postgres psql -d "$DB_NAME" <<EOF
-- Enable pg_trgm extension for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Table: products
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT,
    search tsvector GENERATED ALWAYS AS (
        to_tsvector('english', title || ' ' || author || ' ' || coalesce(description, ''))
    ) STORED
);

-- Table: product_variants
CREATE TABLE IF NOT EXISTS product_variants (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    color TEXT NOT NULL,
    image_path TEXT NOT NULL,
    price NUMERIC(10, 2),
    stock INTEGER NOT NULL CHECK (stock >= 0)
);

-- Table: orders
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL,
    email TEXT NOT NULL,
    products TEXT NOT NULL
    -- You can later add created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Full-text search index
CREATE INDEX IF NOT EXISTS idx_products_search ON products USING GIN(search);

-- Trigram indexes for fuzzy search
CREATE INDEX IF NOT EXISTS trgm_idx_title ON products USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_author ON products USING GIN (author gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_color ON product_variants USING GIN (color gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_orders_email ON orders USING GIN (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_orders_number ON orders USING GIN (CAST(number AS TEXT) gin_trgm_ops);

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
EOF

echo "✅ PostgreSQL database and tables are set up."
