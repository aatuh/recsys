ALTER TABLE items
    ADD COLUMN brand text,
    ADD COLUMN category text,
    ADD COLUMN category_path text[] DEFAULT '{}'::text[],
    ADD COLUMN description text,
    ADD COLUMN image_url text,
    ADD COLUMN metadata_version text;
