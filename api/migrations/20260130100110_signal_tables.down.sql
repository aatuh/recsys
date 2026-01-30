-- 090_signal_tables.sql (down)
DROP TRIGGER IF EXISTS item_popularity_daily_set_updated_at ON item_popularity_daily;
DROP TABLE IF EXISTS item_popularity_daily;

DROP TRIGGER IF EXISTS item_tags_set_updated_at ON item_tags;
DROP TABLE IF EXISTS item_tags;
