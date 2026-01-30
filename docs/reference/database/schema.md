# Database schema (lean)

Use Postgres for:
- tenant config/rules (versioned + current pointers)
- exposure logging
- audit log
- signal tables (DB-only mode): item_tags, item_popularity_daily, item_covisit_daily

Partition high-volume exposure tables by time.

DB-only seed examples:
- `reference/database/db-only-seeding.md`
