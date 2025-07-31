#!/bin/bash

# --- CONFIG ---
DB_NAME="bookmaker"
DB_USER="bookuser"
DB_PASS="securepassword"  # Change this for production!

# Install PostgreSQL (if not already installed)
if ! command -v psql &> /dev/null; then
    echo "Installing PostgreSQL..."
    sudo apt update
    sudo apt install -y postgresql postgresql-contrib
fi

# Start and enable PostgreSQL
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Function to check if DB exists
db_exists() {
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='$1'" | grep -q 1
}

# Function to check if user exists
user_exists() {
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='$1'" | grep -q 1
}

# Create DB if not exists
if db_exists "$DB_NAME"; then
  echo "Database '$DB_NAME' already exists"
else
  echo "Creating database '$DB_NAME'..."
  sudo -u postgres createdb "$DB_NAME"
fi

# Create user if not exists
if user_exists "$DB_USER"; then
  echo "User '$DB_USER' already exists"
else
  echo "Creating user '$DB_USER'..."
  sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';"
fi

# Grant privileges on DB
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# Run schema and extensions setup
sudo -u postgres psql -d "$DB_NAME" <<EOF
-- Enable pg_trgm extension for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create table if it doesn't exist
CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    price NUMERIC(10, 2),
    description TEXT,
    search tsvector GENERATED ALWAYS AS (
        to_tsvector('english', title || ' ' || author || ' ' || coalesce(description, ''))
    ) STORED
);

-- Create full-text search index if not exists
CREATE INDEX IF NOT EXISTS idx_books_search ON books USING GIN(search);

-- Create trigram indexes for fuzzy search
CREATE INDEX IF NOT EXISTS trgm_idx_title ON books USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_author ON books USING GIN (author gin_trgm_ops);

-- Grant permissions on tables and sequences to user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT USAGE, SELECT, UPDATE ON SEQUENCE books_id_seq TO $DB_USER;
EOF

echo "✅ PostgreSQL setup complete."
