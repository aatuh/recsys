-- Create public schema if it doesn't exist.
CREATE SCHEMA IF NOT EXISTS "public";

-- Runs only when /var/lib/postgresql/data is empty.
CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;
