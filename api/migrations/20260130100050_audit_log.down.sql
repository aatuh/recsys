-- 040_audit_log.sql (down)
DROP TRIGGER IF EXISTS audit_log_immutable_ud ON audit_log;
DROP TABLE IF EXISTS audit_log;
