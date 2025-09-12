-- Create public schema if it doesn't exist.
CREATE SCHEMA IF NOT EXISTS "public";

-- Runs only when /var/lib/postgresql/data is empty.
CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;

-- Create "event_type_config" table
CREATE TABLE "public"."event_type_config" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "type" smallint NOT NULL,
  "name" text NULL,
  "weight" double precision NOT NULL,
  "half_life_days" double precision NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("org_id", "namespace", "type"),
  CONSTRAINT "event_type_config_hl_positive" CHECK ((half_life_days IS NULL) OR (half_life_days > (0)::double precision)),
  CONSTRAINT "event_type_config_type_nonneg" CHECK (type >= 0),
  CONSTRAINT "event_type_config_weight_positive" CHECK (weight > (0)::double precision)
);
-- Create index "event_type_config_org_ns_active_idx" to table: "event_type_config"
CREATE INDEX "event_type_config_org_ns_active_idx" ON "public"."event_type_config" ("org_id", "namespace") WHERE (is_active = true);
-- Create "event_type_defaults" table
CREATE TABLE "public"."event_type_defaults" (
  "type" smallint NOT NULL,
  "name" text NOT NULL,
  "weight" double precision NOT NULL,
  "half_life_days" double precision NULL,
  PRIMARY KEY ("type"),
  CONSTRAINT "event_type_defaults_hl_positive" CHECK ((half_life_days IS NULL) OR (half_life_days > (0)::double precision)),
  CONSTRAINT "event_type_defaults_weight_positive" CHECK (weight > (0)::double precision)
);
-- Create "events" table
CREATE TABLE "public"."events" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "user_id" text NOT NULL,
  "item_id" text NOT NULL,
  "type" smallint NOT NULL,
  "value" double precision NOT NULL DEFAULT 1,
  "ts" timestamptz NOT NULL,
  "meta" jsonb NOT NULL DEFAULT '{}',
  "source_event_id" text NULL
);
-- Create index "events_org_ns_item_ts_idx" to table: "events"
CREATE INDEX "events_org_ns_item_ts_idx" ON "public"."events" ("org_id", "namespace", "item_id", "ts");
-- Create index "events_org_ns_ts_idx" to table: "events"
CREATE INDEX "events_org_ns_ts_idx" ON "public"."events" ("org_id", "namespace", "ts");
-- Create index "events_org_ns_user_ts_idx" to table: "events"
CREATE INDEX "events_org_ns_user_ts_idx" ON "public"."events" ("org_id", "namespace", "user_id", "ts");
-- Create index "events_source_uidx" to table: "events"
CREATE UNIQUE INDEX "events_source_uidx" ON "public"."events" ("org_id", "namespace", "source_event_id");
-- Create "items" table
CREATE TABLE "public"."items" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "item_id" text NOT NULL,
  "available" boolean NOT NULL DEFAULT true,
  "price" numeric(12,2) NULL,
  "tags" text[] NOT NULL DEFAULT '{}',
  "embedding" public.vector(384) NULL,
  "props" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("org_id", "namespace", "item_id")
);
-- Create index "items_ns_created_idx" to table: "items"
CREATE INDEX "items_ns_created_idx" ON "public"."items" ("org_id", "namespace", "created_at");
-- Create index "items_org_ns_available_item_idx" to table: "items"
CREATE INDEX "items_org_ns_available_item_idx" ON "public"."items" ("org_id", "namespace", "available", "item_id");
-- Create index "items_tags_gin_idx" to table: "items"
CREATE INDEX "items_tags_gin_idx" ON "public"."items" USING gin ("tags");
-- Set comment to column: "embedding" on table: "items"
COMMENT ON COLUMN "public"."items"."embedding" IS 'Text embedding for ANN similarity';
-- Create "organizations" table
CREATE TABLE "public"."organizations" (
  "org_id" uuid NOT NULL,
  "name" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("org_id")
);
-- Create "users" table
CREATE TABLE "public"."users" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "user_id" text NOT NULL,
  "traits" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("org_id", "namespace", "user_id")
);
-- Create index "users_ns_created_idx" to table: "users"
CREATE INDEX "users_ns_created_idx" ON "public"."users" ("org_id", "namespace", "created_at");
-- Create "namespaces" table
CREATE TABLE "public"."namespaces" (
  "id" uuid NOT NULL,
  "org_id" uuid NOT NULL,
  "name" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "namespaces_org_fkey" FOREIGN KEY ("org_id") REFERENCES "public"."organizations" ("org_id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "namespaces_org_id_name_uq" to table: "namespaces"
CREATE UNIQUE INDEX "namespaces_org_id_name_uq" ON "public"."namespaces" ("org_id", "name");

-- Insert default event types if they don't exist
INSERT INTO event_type_defaults(type, name, weight, half_life_days) VALUES
  (0, 'view', 0.1, NULL),
  (1, 'click', 0.3, NULL),
  (2, 'add', 0.7, NULL),
  (3, 'purchase', 1.0, NULL),
  (4, 'custom', 0.2, NULL)
ON CONFLICT (type) DO UPDATE
  SET name = EXCLUDED.name,
      weight = EXCLUDED.weight,
      half_life_days = EXCLUDED.half_life_days;
