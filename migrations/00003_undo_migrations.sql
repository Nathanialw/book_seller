-- Combined Migrations Version: 00003
-- This file contains all migrations for this version

-- Undo migration for table: tests
ALTER TABLE tests ADD COLUMN IF NOT EXISTS ImagePath VARCHAR;
