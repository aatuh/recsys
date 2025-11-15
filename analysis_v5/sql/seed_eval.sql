\set ON_ERROR_STOP on

BEGIN;

-- Organization + namespace bootstrap -------------------------------------------------------------
INSERT INTO organizations (org_id, name)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'Evaluation Control Org')
ON CONFLICT (org_id) DO UPDATE
SET name = EXCLUDED.name;

INSERT INTO namespaces (id, org_id, name)
VALUES
  ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000001', 'retail_en_us'),
  ('22222222-2222-2222-2222-222222222222', '00000000-0000-0000-0000-000000000001', 'media_fi_fi')
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name;

-- Clean slate for the namespaces under test -----------------------------------------------------
DELETE FROM events
WHERE org_id = '00000000-0000-0000-0000-000000000001'
  AND namespace IN ('retail_en_us', 'media_fi_fi');

DELETE FROM users
WHERE org_id = '00000000-0000-0000-0000-000000000001'
  AND namespace IN ('retail_en_us', 'media_fi_fi');

DELETE FROM items
WHERE org_id = '00000000-0000-0000-0000-000000000001'
  AND namespace IN ('retail_en_us', 'media_fi_fi');

-- Catalog inventory -----------------------------------------------------------------------------
WITH item_data (namespace, item_id, category, brand, price, tags, props) AS (
  VALUES
    ('retail_en_us', 'sku_retail_electronics_001', 'Electronics', 'Voltify', 349.00,
      ARRAY['electronics','smart','brand:voltify','category:electronics','domain:retail','locale:en-us'],
      '{"margin":0.32,"novelty":0.44,"popularity_hint":0.91,"popularity_rank_norm":0.15,"season":"holiday","inventory":"healthy"}'::jsonb),
    ('retail_en_us', 'sku_retail_electronics_002', 'Electronics', 'Nimbus', 499.00,
      ARRAY['electronics','premium','brand:nimbus','category:electronics','domain:retail','locale:en-us'],
      '{"margin":0.27,"novelty":0.58,"popularity_hint":0.72,"popularity_rank_norm":0.25,"season":"holiday","inventory":"balanced"}'::jsonb),
    ('retail_en_us', 'sku_retail_books_001', 'Books', 'LeafPress', 24.00,
      ARRAY['books','reading','brand:leafpress','category:books','domain:retail','locale:en-us'],
      '{"margin":0.41,"novelty":0.35,"popularity_hint":0.64,"popularity_rank_norm":0.43,"season":"back_to_school","inventory":"healthy"}'::jsonb),
    ('retail_en_us', 'sku_retail_books_002', 'Books', 'Quanta', 19.00,
      ARRAY['books','long_tail','brand:quanta','category:books','domain:retail','locale:en-us'],
      '{"margin":0.46,"novelty":0.61,"popularity_hint":0.28,"popularity_rank_norm":0.82,"season":"evergreen","inventory":"ample"}'::jsonb),
    ('retail_en_us', 'sku_retail_home_001', 'Home', 'CozyNest', 129.00,
      ARRAY['home','decor','brand:cozynest','category:home','domain:retail','locale:en-us'],
      '{"margin":0.38,"novelty":0.33,"popularity_hint":0.69,"popularity_rank_norm":0.36,"season":"fall","inventory":"healthy"}'::jsonb),
    ('retail_en_us', 'sku_retail_home_002', 'Home', 'BrightLiving', 219.00,
      ARRAY['home','smart','brand:brightliving','category:home','domain:retail','locale:en-us'],
      '{"margin":0.29,"novelty":0.67,"popularity_hint":0.41,"popularity_rank_norm":0.55,"season":"holiday","inventory":"tight"}'::jsonb),
    ('retail_en_us', 'sku_retail_fitness_001', 'Fitness', 'PulseGear', 159.00,
      ARRAY['fitness','wearable','brand:pulsegear','category:fitness','domain:retail','locale:en-us'],
      '{"margin":0.33,"novelty":0.48,"popularity_hint":0.76,"popularity_rank_norm":0.31,"season":"new_year","inventory":"healthy"}'::jsonb),
    ('retail_en_us', 'sku_retail_fashion_001', 'Fashion', 'AuraThreads', 84.00,
      ARRAY['fashion','style','brand:aurathreads','category:fashion','domain:retail','locale:en-us'],
      '{"margin":0.49,"novelty":0.71,"popularity_hint":0.37,"popularity_rank_norm":0.61,"season":"fall","inventory":"balanced"}'::jsonb),
    ('retail_en_us', 'sku_retail_gourmet_001', 'Gourmet', 'Epicurean', 34.00,
      ARRAY['gourmet','kitchen','brand:epicurean','category:gourmet','domain:retail','locale:en-us'],
      '{"margin":0.36,"novelty":0.52,"popularity_hint":0.54,"popularity_rank_norm":0.49,"season":"holiday","inventory":"ample"}'::jsonb),
    ('retail_en_us', 'sku_retail_outdoors_001', 'Outdoors', 'SummitPeak', 289.00,
      ARRAY['outdoors','adventure','brand:summitpeak','category:outdoors','domain:retail','locale:en-us'],
      '{"margin":0.31,"novelty":0.63,"popularity_hint":0.47,"popularity_rank_norm":0.57,"season":"summer","inventory":"healthy"}'::jsonb),
    ('retail_en_us', 'sku_retail_longtail_001', 'Fashion', 'IndieWeave', 59.00,
      ARRAY['fashion','artisan','long_tail','brand:indieweave','category:fashion','domain:retail','locale:en-us'],
      '{"margin":0.52,"novelty":0.83,"popularity_hint":0.12,"popularity_rank_norm":0.93,"season":"evergreen","inventory":"small_batch"}'::jsonb),
    ('retail_en_us', 'sku_retail_newdrop_001', 'Electronics', 'Voltify', 189.00,
      ARRAY['electronics','new_drop','brand:voltify','category:electronics','domain:retail','locale:en-us'],
      '{"margin":0.35,"novelty":0.89,"popularity_hint":0.18,"popularity_rank_norm":0.88,"season":"launch_week","inventory":"tight"}'::jsonb),
    ('media_fi_fi', 'vid_media_news_001', 'News', 'NordicSignal', NULL,
      ARRAY['media','news','locale:fi-fi','domain:media','category:news','language:fi'],
      '{"duration_minutes":18,"novelty":0.37,"popularity_hint":0.78,"popularity_rank_norm":0.29,"content_rating":"G"}'::jsonb),
    ('media_fi_fi', 'vid_media_news_002', 'News', 'NordicSignal', NULL,
      ARRAY['media','news','long_tail','locale:fi-fi','domain:media','category:news'],
      '{"duration_minutes":12,"novelty":0.58,"popularity_hint":0.33,"popularity_rank_norm":0.63,"content_rating":"G"}'::jsonb),
    ('media_fi_fi', 'vid_media_drama_001', 'Drama', 'AuroraShows', NULL,
      ARRAY['media','drama','locale:fi-fi','domain:media','category:drama'],
      '{"duration_minutes":52,"novelty":0.46,"popularity_hint":0.69,"popularity_rank_norm":0.34,"content_rating":"PG-13"}'::jsonb),
    ('media_fi_fi', 'vid_media_drama_002', 'Drama', 'AuroraShows', NULL,
      ARRAY['media','drama','long_tail','locale:fi-fi','domain:media','category:drama'],
      '{"duration_minutes":48,"novelty":0.73,"popularity_hint":0.24,"popularity_rank_norm":0.79,"content_rating":"PG"}'::jsonb),
    ('media_fi_fi', 'vid_media_doc_001', 'Documentary', 'NorthDocs', NULL,
      ARRAY['media','documentary','locale:fi-fi','domain:media','category:documentary'],
      '{"duration_minutes":65,"novelty":0.51,"popularity_hint":0.57,"popularity_rank_norm":0.46,"content_rating":"PG"}'::jsonb),
    ('media_fi_fi', 'vid_media_sports_001', 'Sports', 'SuomiPlay', NULL,
      ARRAY['media','sports','locale:fi-fi','domain:media','category:sports'],
      '{"duration_minutes":95,"novelty":0.44,"popularity_hint":0.83,"popularity_rank_norm":0.22,"content_rating":"G"}'::jsonb),
    ('media_fi_fi', 'vid_media_sports_002', 'Sports', 'SuomiPlay', NULL,
      ARRAY['media','sports','long_tail','locale:fi-fi','domain:media','category:sports'],
      '{"duration_minutes":88,"novelty":0.66,"popularity_hint":0.31,"popularity_rank_norm":0.67,"content_rating":"G"}'::jsonb),
    ('media_fi_fi', 'vid_media_kids_001', 'Kids', 'BrightNordic', NULL,
      ARRAY['media','kids','locale:fi-fi','domain:media','category:kids'],
      '{"duration_minutes":26,"novelty":0.59,"popularity_hint":0.62,"popularity_rank_norm":0.41,"content_rating":"G"}'::jsonb),
    ('media_fi_fi', 'vid_media_newdrop_001', 'News', 'NordicSignal', NULL,
      ARRAY['media','news','breaking','locale:fi-fi','domain:media','category:news'],
      '{"duration_minutes":9,"novelty":0.91,"popularity_hint":0.19,"popularity_rank_norm":0.9,"content_rating":"G"}'::jsonb)
)
INSERT INTO items (org_id, namespace, item_id, available, price, tags, props, created_at, updated_at)
SELECT
  '00000000-0000-0000-0000-000000000001',
  namespace,
  item_id,
  TRUE,
  price,
  tags,
  props || jsonb_build_object(
    'brand', brand,
    'category', category,
    'domain', CASE WHEN namespace = 'retail_en_us' THEN 'retail' ELSE 'media' END,
    'locale', CASE WHEN namespace = 'retail_en_us' THEN 'en-US' ELSE 'fi-FI' END
  ),
  '2025-09-01T00:00:00Z'::timestamptz,
  '2025-10-01T00:00:00Z'::timestamptz
FROM item_data
ON CONFLICT (org_id, namespace, item_id) DO UPDATE
SET
  available = EXCLUDED.available,
  price = EXCLUDED.price,
  tags = EXCLUDED.tags,
  props = EXCLUDED.props,
  updated_at = now();

-- User fixtures ---------------------------------------------------------------------------------
WITH user_data (namespace, user_id, traits) AS (
  VALUES
    ('retail_en_us', 'retail_power_001',
      '{"segment":"power_users","locale":"en-US","region":"us-east","traffic_tier":"100","device_mix":["ios","web"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_power_002',
      '{"segment":"power_users","locale":"en-US","region":"us-west","traffic_tier":"100","device_mix":["android","web"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_trend_001',
      '{"segment":"trend_seekers","locale":"en-US","region":"us-central","traffic_tier":"10","device_mix":["web"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_trend_002',
      '{"segment":"trend_seekers","locale":"en-US","region":"us-south","traffic_tier":"10","device_mix":["ios"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_niche_001',
      '{"segment":"niche_readers","locale":"en-US","region":"us-northeast","traffic_tier":"10","device_mix":["web"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_niche_002',
      '{"segment":"niche_readers","locale":"en-US","region":"us-midwest","traffic_tier":"10","device_mix":["android"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_weekend_001',
      '{"segment":"weekend_adventurers","locale":"en-US","region":"us-rockies","traffic_tier":"10","device_mix":["ios","web"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_weekend_002',
      '{"segment":"weekend_adventurers","locale":"en-US","region":"us-southwest","traffic_tier":"10","device_mix":["android"],"lifecycle":"existing"}'::jsonb),
    ('retail_en_us', 'retail_zero_001',
      '{"segment":"new_users","locale":"en-US","region":"us-midatlantic","traffic_tier":"10","device_mix":["ios"],"lifecycle":"new"}'::jsonb),
    ('retail_en_us', 'retail_zero_002',
      '{"segment":"new_users","locale":"en-US","region":"us-gulf","traffic_tier":"10","device_mix":["web"],"lifecycle":"new"}'::jsonb),
    ('media_fi_fi', 'media_news_001',
      '{"segment":"news_pro","locale":"fi-FI","region":"fi-south","traffic_tier":"100","device_mix":["web"],"lifecycle":"existing"}'::jsonb),
    ('media_fi_fi', 'media_news_002',
      '{"segment":"news_pro","locale":"fi-FI","region":"fi-west","traffic_tier":"10","device_mix":["ios"],"lifecycle":"existing"}'::jsonb),
    ('media_fi_fi', 'media_sports_001',
      '{"segment":"sports_streamers","locale":"fi-FI","region":"fi-central","traffic_tier":"100","device_mix":["tv"],"lifecycle":"existing"}'::jsonb),
    ('media_fi_fi', 'media_sports_002',
      '{"segment":"sports_streamers","locale":"fi-FI","region":"fi-south","traffic_tier":"10","device_mix":["android"],"lifecycle":"existing"}'::jsonb),
    ('media_fi_fi', 'media_drama_001',
      '{"segment":"story_seekers","locale":"fi-FI","region":"fi-east","traffic_tier":"10","device_mix":["web"],"lifecycle":"existing"}'::jsonb),
    ('media_fi_fi', 'media_family_001',
      '{"segment":"family_profiles","locale":"fi-FI","region":"fi-southeast","traffic_tier":"10","device_mix":["tv"],"lifecycle":"existing"}'::jsonb),
    ('media_fi_fi', 'media_zero_001',
      '{"segment":"onboard_trial","locale":"fi-FI","region":"fi-north","traffic_tier":"10","device_mix":["ios"],"lifecycle":"new"}'::jsonb)
)
INSERT INTO users (org_id, namespace, user_id, traits, created_at, updated_at)
SELECT
  '00000000-0000-0000-0000-000000000001',
  namespace,
  user_id,
  traits,
  '2025-09-05T00:00:00Z'::timestamptz,
  '2025-10-01T00:00:00Z'::timestamptz
FROM user_data
ON CONFLICT (org_id, namespace, user_id) DO UPDATE
SET
  traits = EXCLUDED.traits,
  updated_at = now();

COMMIT;

-- Event generation (executed after transaction to keep DO blocks simple) -------------------------
DO $$
DECLARE
  org CONSTANT uuid := '00000000-0000-0000-0000-000000000001';
  base_ts CONSTANT timestamptz := '2025-09-01 08:00:00+00'::timestamptz;
  retail_users text[] := ARRAY[
    'retail_power_001','retail_power_002','retail_trend_001','retail_trend_002',
    'retail_niche_001','retail_niche_002','retail_weekend_001','retail_weekend_002'
  ];
  retail_segments text[] := ARRAY[
    'power_users','power_users','trend_seekers','trend_seekers',
    'niche_readers','niche_readers','weekend_adventurers','weekend_adventurers'
  ];
  retail_items text[] := ARRAY[
    'sku_retail_electronics_001','sku_retail_electronics_002','sku_retail_books_001',
    'sku_retail_books_002','sku_retail_home_001','sku_retail_home_002',
    'sku_retail_fitness_001','sku_retail_fashion_001','sku_retail_gourmet_001',
    'sku_retail_outdoors_001','sku_retail_longtail_001','sku_retail_newdrop_001'
  ];
  event_pattern smallint[] := ARRAY[0,0,1,0,2,0,1,0,3];
  user_idx int;
  event_idx int;
  event_type smallint;
  chosen_item text;
  segment text;
BEGIN
  FOR user_idx IN 1..array_length(retail_users, 1) LOOP
    segment := retail_segments[user_idx];
    FOR event_idx IN 1..28 LOOP
      chosen_item := retail_items[((event_idx + user_idx - 2) % array_length(retail_items, 1)) + 1];
      event_type := event_pattern[((event_idx - 1) % array_length(event_pattern, 1)) + 1];
      INSERT INTO events (org_id, namespace, user_id, item_id, type, value, ts, meta)
      VALUES (
        org,
        'retail_en_us',
        retail_users[user_idx],
        chosen_item,
        event_type,
        CASE WHEN event_type = 3 THEN 1 ELSE 1 END,
        base_ts + ((user_idx - 1) * 12 + event_idx) * INTERVAL '2 hours',
        jsonb_build_object(
          'segment', segment,
          'surface', CASE WHEN event_idx % 3 = 0 THEN 'cart' WHEN event_idx % 2 = 0 THEN 'home' ELSE 'pdp' END,
          'device', CASE WHEN event_idx % 2 = 0 THEN 'ios' ELSE 'web' END,
          'locale', 'en-US',
          'traffic_tier', CASE WHEN user_idx <= 2 THEN '100' ELSE '10' END
        )
      );
    END LOOP;
  END LOOP;
END $$;

DO $$
DECLARE
  org CONSTANT uuid := '00000000-0000-0000-0000-000000000001';
  base_ts CONSTANT timestamptz := '2025-09-03 06:00:00+00'::timestamptz;
  media_users text[] := ARRAY[
    'media_news_001','media_news_002','media_sports_001',
    'media_sports_002','media_drama_001','media_family_001'
  ];
  media_segments text[] := ARRAY[
    'news_pro','news_pro','sports_streamers',
    'sports_streamers','story_seekers','family_profiles'
  ];
  media_items text[] := ARRAY[
    'vid_media_news_001','vid_media_news_002','vid_media_drama_001','vid_media_drama_002',
    'vid_media_doc_001','vid_media_sports_001','vid_media_sports_002','vid_media_kids_001','vid_media_newdrop_001'
  ];
  event_pattern smallint[] := ARRAY[0,0,1,0,0,1,3];
  user_idx int;
  event_idx int;
  event_type smallint;
  chosen_item text;
  segment text;
BEGIN
  FOR user_idx IN 1..array_length(media_users, 1) LOOP
    segment := media_segments[user_idx];
    FOR event_idx IN 1..24 LOOP
      chosen_item := media_items[((event_idx + user_idx - 2) % array_length(media_items, 1)) + 1];
      event_type := event_pattern[((event_idx - 1) % array_length(event_pattern, 1)) + 1];
      INSERT INTO events (org_id, namespace, user_id, item_id, type, value, ts, meta)
      VALUES (
        org,
        'media_fi_fi',
        media_users[user_idx],
        chosen_item,
        event_type,
        1,
        base_ts + ((user_idx - 1) * 10 + event_idx) * INTERVAL '3 hours',
        jsonb_build_object(
          'segment', segment,
          'surface', CASE WHEN event_idx % 4 = 0 THEN 'continue_watching' ELSE 'home' END,
          'device', CASE WHEN user_idx <= 2 THEN 'web' WHEN user_idx <= 4 THEN 'tv' ELSE 'ios' END,
          'locale', 'fi-FI',
          'traffic_tier', CASE WHEN user_idx <= 3 THEN '100' ELSE '10' END
        )
      );
    END LOOP;
  END LOOP;
END $$;

-- cold-start placeholder interactions -----------------------------------------------------------
INSERT INTO events (org_id, namespace, user_id, item_id, type, value, ts, meta)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'retail_en_us', 'retail_zero_001', 'sku_retail_newdrop_001', 0, 1,
    '2025-10-01T10:00:00Z'::timestamptz, '{"segment":"new_users","surface":"home","device":"ios","locale":"en-US"}'),
  ('00000000-0000-0000-0000-000000000001', 'media_fi_fi', 'media_zero_001', 'vid_media_newdrop_001', 0, 1,
    '2025-10-01T12:00:00Z'::timestamptz, '{"segment":"onboard_trial","surface":"home","device":"ios","locale":"fi-FI"}')
ON CONFLICT DO NOTHING;
