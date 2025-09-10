-- Remove the user_id column from tenants table
ALTER TABLE tenants DROP COLUMN IF EXISTS user_id;

-- Drop tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
