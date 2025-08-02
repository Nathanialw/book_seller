# --- CONFIG ---
$DB_NAME = "bookmaker"
$DB_USER = "bookuser"
# TODO: Change this for production!
$DB_PASS = "securepassword"

# PostgreSQL bin path (adjust if needed)
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

# Ask for the postgres password
$PostgresPass = Read-Host -Prompt "Enter password for postgres user" -AsSecureString
$BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($PostgresPass)
$PostgresPassPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringBSTR($BSTR)

# Set PGPASSWORD env var for non-interactive use
$env:PGPASSWORD = $PostgresPassPlain

# Create DB if not exists
if (Db-Exists) {
    Write-Host "Database '$DB_NAME' already exists"
} else {
    Write-Host "Creating database '$DB_NAME'..."
    createdb -U postgres $DB_NAME
}

# Create user if not exists
if (User-Exists) {
    Write-Host "User '$DB_USER' already exists"
} else {
    Write-Host "Creating user '$DB_USER'..."
    psql -U postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';"
}

# Grant privileges on DB
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# Set up schema and extensions
$setupSql = @"
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT,
    search tsvector GENERATED ALWAYS AS (
        to_tsvector('english', title || ' ' || author || ' ' || coalesce(description, ''))
    ) STORED
);

CREATE TABLE IF NOT EXISTS book_variants (
    id SERIAL PRIMARY KEY,
    book_id INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    color TEXT NOT NULL,
    price NUMERIC(10, 2),
    image_path TEXT NOT NULL,
    stock INTEGER NOT NULL CHECK (stock >= 0)
);

CREATE INDEX IF NOT EXISTS idx_books_search ON books USING GIN(search);

CREATE INDEX IF NOT EXISTS trgm_idx_title ON books USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_author ON books USING GIN (author gin_trgm_ops);
CREATE INDEX IF NOT EXISTS trgm_idx_color ON book_variants USING GIN (color gin_trgm_ops);

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT USAGE, SELECT, UPDATE ON SEQUENCE books_id_seq TO $DB_USER;
GRANT USAGE, SELECT, UPDATE ON SEQUENCE book_variants_id_seq TO $DB_USER;
"@

# Write the SQL to temp file and run it
$tempSqlFile = "$env:TEMP\init_db.sql"
$setupSql | Out-File -Encoding UTF8 $tempSqlFile
psql -U postgres -d $DB_NAME -f $tempSqlFile
Remove-Item $tempSqlFile

Write-Host "✅ PostgreSQL setup complete on Windows."
