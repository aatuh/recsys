-- 070_partition_helpers.sql (down)
DROP FUNCTION IF EXISTS ensure_monthly_partitions(regclass, date, integer);
