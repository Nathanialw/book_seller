#!/bin/bash

DB_NAME="ecommerce"
DB_USER="admin"
DB_PASS="securepassword"

db_exists() {
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='$1'" | grep -q 1
}

user_exists() {
  sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='$1'" | grep -q 1
}

if db_exists "$DB_NAME"; then
  echo "✅ Database '$DB_NAME' already exists."
else
  echo "🔧 Creating database '$DB_NAME'..."
  sudo -u postgres createdb "$DB_NAME"
fi

if user_exists "$DB_USER"; then
  echo "✅ User '$DB_USER' already exists."
else
  echo "🔧 Creating user '$DB_USER'..."
  sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';"
fi

sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
