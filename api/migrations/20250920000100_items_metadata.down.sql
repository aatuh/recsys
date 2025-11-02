ALTER TABLE items
    DROP COLUMN IF EXISTS metadata_version,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS category_path,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS brand;
