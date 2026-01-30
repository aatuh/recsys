-- 070_partition_helpers.sql
-- Purpose: helper to create monthly partitions ahead of time for a parent
-- partitioned table. Run this via a scheduled job / admin process.

CREATE OR REPLACE FUNCTION ensure_monthly_partitions(
  parent_table regclass,
  start_month  date,
  months_ahead integer
) RETURNS void AS $$
DECLARE
  i integer;
  from_ts timestamptz;
  to_ts   timestamptz;
  part_name text;
BEGIN
  IF months_ahead < 1 THEN
    RAISE EXCEPTION 'months_ahead must be >= 1';
  END IF;

  FOR i IN 0..(months_ahead - 1) LOOP
    from_ts := (date_trunc('month', start_month)::date + (i || ' months')::interval)
               ::timestamptz;
    to_ts   := (from_ts + interval '1 month');

    part_name := format('%s_%s',
      replace(parent_table::text, '.', '_'),
      to_char(from_ts, 'YYYY_MM')
    );

    EXECUTE format(
      'CREATE TABLE IF NOT EXISTS %I PARTITION OF %s
       FOR VALUES FROM (%L) TO (%L)',
      part_name, parent_table::text, from_ts, to_ts
    );
  END LOOP;
END;
$$ LANGUAGE plpgsql;
