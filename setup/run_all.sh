#!/bin/bash
set -e  # Exit on any error

# --- CONFIG ---
DB_NAME="ecommerce"

FRAMEWORK_DIR="$(cd "$(dirname "$0")" && pwd)"
EXTENDS_DIR="$FRAMEWORK_DIR/../schema_extends"

# --- Ensure project schema_extends folder exists ---
if [ ! -d "$EXTENDS_DIR" ]; then
    mkdir -p "$EXTENDS_DIR"
    echo "📂 Created project schema extensions folder: $EXTENDS_DIR"

    # Create .gitignore to protect this folder from being committed to framework repo
    cat > "$EXTENDS_DIR/.gitignore" <<EOF
# Ignore everything in this folder
*
# But keep this file
!.gitignore
EOF
    echo "🛡️ Added .gitignore to protect project schema extensions."

    # Create README for instructions
    cat > "$EXTENDS_DIR/README.txt" <<EOF
This folder stores project-specific SQL schema extensions for your eCommerce site.

- Place any .sql files here to extend the base schema without modifying the framework.
- Files will be run automatically by run_all.sh after the base schema.
- Example: Add new columns to products table, create new indexes, etc.

Example file:
ALTER TABLE products ADD COLUMN IF NOT EXISTS isbn TEXT;
EOF
    echo "ℹ️ Added README.txt with instructions."
fi

echo "==> Installing PostgreSQL if needed..."
bash "$FRAMEWORK_DIR/install_postgres.sh"

echo "==> Creating database and user..."
bash "$FRAMEWORK_DIR/create_db_and_user.sh"

echo "==> Applying base schema (framework)..."
if [ -d "$FRAMEWORK_DIR/schema" ]; then
    for file in "$FRAMEWORK_DIR/schema"/*.sql; do
        echo "----> Applying $file"
        sudo -u postgres psql -d "$DB_NAME" -f "$file"
    done
else
    echo "⚠️ No base schema folder found ($FRAMEWORK_DIR/schema)."
fi

echo "==> Applying project-specific schema extensions..."
if [ "$(ls -A "$EXTENDS_DIR"/*.sql 2>/dev/null)" ]; then
    for file in "$EXTENDS_DIR"/*.sql; do
        echo "----> Applying $file"
        sudo -u postgres psql -d "$DB_NAME" -f "$file"
    done
else
    echo "ℹ️ No .sql files found in $EXTENDS_DIR."
fi

echo "✅ All setup steps completed successfully!"
