# --- CONFIG ---
$DB_NAME = "ecommerce"
$DB_USER = "admin"
$DB_PASS = "securepassword"  # ⚠️ Change this for production!

# PostgreSQL bin path (adjust this if needed)
$PG_BIN = "C:\Program Files\PostgreSQL\17\bin"
$env:Path += ";$PG_BIN"

# Function to check if PostgreSQL tools are available
function Check-PostgresInstalled {
    if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
        Write-Error "PostgreSQL tools not found in PATH. Please install PostgreSQL and ensure psql is in your system PATH."
        exit 1
    }
}

# Function to check if DB exists
function Db-Exists {
    $result = psql -U postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'"
    return $result.Trim() -eq "1"
}

# Function to check if user exists
function User-Exists {
    $result = psql -U postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'"
    return $result.Trim() -eq "1"
}

# Ensure PostgreSQL is installed
Check-PostgresInstalled

# Ask for the postgres password securely
$PostgresPass = Read-Host -Prompt "Enter password for postgres user" -AsSecureString
$BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($PostgresPass)
$PostgresPassPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringBSTR($BSTR)

# Set env var for non-interactive psql
$env:PGPASSWORD = $PostgresPassPlain

# Create DB if not exists
if (Db-Exists) {
    Write-Host "✅ Database '$DB_NAME' already exists."
} else {
    Write-Host "🔧 Creating database '$DB_NAME'..."
    createdb -U postgres $DB_NAME
}

# Create user if not exists
if (User-Exists) {
    Write-Host "✅ User '$DB_USER' already exists."
} else {
    Write-Host "🔧 Creating user '$DB_USER'..."
    psql -U postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';"
}

# Grant privileges on DB
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# Set up schema and extensions
$setupSql = @"
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT,
    search tsvector GENERATED ALWAYS AS (
        to_tsvector('english', title || ' ' || author || ' ' || coalesce(description, ''))
    ) STORED
);

CREATE TABLE IF NOT EXISTS product_variants (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    color TEXT NOT NULL,
    image_path TEXT NOT NULL,
    price NUMERIC(10, 2),
    stock INTEGER NOT NULL CHECK (stock >= 0)
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    number INTEGER NOT NULL,
    email TEXT NOT NULL,
    products TEXT NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_products_search ON products USING GIN(search);
CREATE INDEX IF NOT EXISTS trgm_idx_title ON products USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_author ON products USING GIN (author gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_color ON product_variants USING GIN (color gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_orders_email ON orders USING GIN (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_orders_number ON orders USING GIN (CAST(number AS TEXT) gin_trgm_ops);

-- Permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
"@

# Run SQL setup
$tempSqlFile = "$env:TEMP\init_ecommerce_db.sql"
$setupSql | Set-Content -Encoding UTF8 $tempSqlFile
psql -U postgres -d $DB_NAME -f $tempSqlFile
Remove-Item $tempSqlFile

Write-Host "✅ PostgreSQL setup complete on Windows."
