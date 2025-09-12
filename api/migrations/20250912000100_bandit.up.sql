-- Create bandit_policies table
CREATE TABLE IF NOT EXISTS "public"."bandit_policies" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "policy_id" text NOT NULL,
  "name" text NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "config" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("org_id", "namespace", "policy_id")
);

-- Index for active policies per namespace (partial)
CREATE INDEX IF NOT EXISTS "bandit_policies_active_idx"
  ON "public"."bandit_policies" ("org_id", "namespace")
  WHERE (is_active = true);

-- Create bandit_stats table
CREATE TABLE IF NOT EXISTS "public"."bandit_stats" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "surface" text NOT NULL,
  "bucket_key" text NOT NULL,
  "policy_id" text NOT NULL,
  "algo" text NOT NULL,
  "trials" bigint NOT NULL DEFAULT 0,
  "successes" bigint NOT NULL DEFAULT 0,
  "alpha" double precision NOT NULL DEFAULT 1.0,
  "beta" double precision NOT NULL DEFAULT 1.0,
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("org_id", "namespace", "surface", "bucket_key", "policy_id", "algo")
);

-- Create bandit_decisions_log table
CREATE TABLE IF NOT EXISTS "public"."bandit_decisions_log" (
  "id" bigserial NOT NULL,
  "ts" timestamptz NOT NULL DEFAULT now(),
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "surface" text NOT NULL,
  "bucket_key" text NOT NULL,
  "policy_id" text NOT NULL,
  "algo" text NOT NULL,
  "explore" boolean NOT NULL,
  "request_id" text NULL,
  "meta" jsonb NULL,
  PRIMARY KEY ("id")
);

-- Supporting index for decisions log queries
CREATE INDEX IF NOT EXISTS "bandit_decisions_idx"
  ON "public"."bandit_decisions_log" ("org_id", "namespace", "surface", "bucket_key");

-- Create bandit_rewards_log table
CREATE TABLE IF NOT EXISTS "public"."bandit_rewards_log" (
  "id" bigserial NOT NULL,
  "ts" timestamptz NOT NULL DEFAULT now(),
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "surface" text NOT NULL,
  "bucket_key" text NOT NULL,
  "policy_id" text NOT NULL,
  "algo" text NOT NULL,
  "reward" boolean NOT NULL,
  "request_id" text NULL,
  "meta" jsonb NULL,
  PRIMARY KEY ("id")
);

-- Supporting index for rewards log queries
CREATE INDEX IF NOT EXISTS "bandit_rewards_idx"
  ON "public"."bandit_rewards_log" ("org_id", "namespace", "surface", "bucket_key");


