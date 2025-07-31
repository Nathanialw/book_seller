#!/bin/bash

# --- CONFIG ---
DB_NAME="bookmaker"
DB_USER="bookuser"
DB_PASS="securepassword"  # Change this for production!
PG_VERSION=14             # Change if your system uses a different version

# Install PostgreSQL (if not already installed)
if ! command -v psql &> /dev/null; then
    echo "Installing PostgreSQL..."
    sudo apt update
    sudo apt install -y postgresql postgresql-contrib
fi

# Start and enable PostgreSQL
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Create DB and user
echo "Creating database and user..."
sudo -u postgres psql <<EOF
CREATE DATABASE $DB_NAME;
CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
\c $DB_NAME
CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    price NUMERIC(10, 2),
    description TEXT,
    search tsvector GENERATED ALWAYS AS (
        to_tsvector('english', title || ' ' || author || ' ' || coalesce(description, ''))
    ) STORED
);
CREATE INDEX idx_books_search ON books USING GIN(search);
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT USAGE, SELECT, UPDATE ON SEQUENCE books_id_seq TO $DB_USER;
EOF

echo "✅ PostgreSQL setup complete."
