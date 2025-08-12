-- Combined Migrations Version: 00003
-- This file contains all migrations for this version

-- Migration for table: tests
ALTER TABLE tests DROP COLUMN IF EXISTS ImagePath CASCADE;
