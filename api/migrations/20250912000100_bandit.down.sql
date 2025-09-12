-- Drop indexes first (they are auto-dropped with table, but explicit is ok)
DROP INDEX IF EXISTS "public"."bandit_rewards_idx";
DROP INDEX IF EXISTS "public"."bandit_decisions_idx";
DROP INDEX IF EXISTS "public"."bandit_policies_active_idx";

-- Drop tables
DROP TABLE IF EXISTS "public"."bandit_rewards_log";
DROP TABLE IF EXISTS "public"."bandit_decisions_log";
DROP TABLE IF EXISTS "public"."bandit_stats";
DROP TABLE IF EXISTS "public"."bandit_policies";


