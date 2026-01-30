-- 003_util.sql
-- Purpose: shared helper functions/triggers.

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION prevent_update_delete()
RETURNS TRIGGER AS $$
BEGIN
  RAISE EXCEPTION 'Table % is append-only', TG_TABLE_NAME
    USING ERRCODE = '0A000';
END;
$$ LANGUAGE plpgsql;
