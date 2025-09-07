-- Create "events" table
CREATE TABLE "public"."events" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "user_id" text NOT NULL,
  "item_id" text NOT NULL,
  "type" smallint NOT NULL,
  "value" double precision NOT NULL DEFAULT 1,
  "ts" timestamptz NOT NULL,
  "meta" jsonb NOT NULL DEFAULT '{}'
);
-- Create index "events_org_ns_item_ts_idx" to table: "events"
CREATE INDEX "events_org_ns_item_ts_idx" ON "public"."events" ("org_id", "namespace", "item_id", "ts");
-- Create index "events_org_ns_ts_idx" to table: "events"
CREATE INDEX "events_org_ns_ts_idx" ON "public"."events" ("org_id", "namespace", "ts");
-- Create index "events_org_ns_user_ts_idx" to table: "events"
CREATE INDEX "events_org_ns_user_ts_idx" ON "public"."events" ("org_id", "namespace", "user_id", "ts");
-- Create "items" table
CREATE TABLE "public"."items" (
  "org_id" uuid NOT NULL,
  "namespace" text NOT NULL,
  "item_id" text NOT NULL,
  "available" boolean NOT NULL DEFAULT true,
  "price" numeric(12,2) NULL,
  "tags" text[] NOT NULL DEFAULT '{}',
  "vector" real[] NULL,
  "props" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("org_id", "namespace", "item_id")
);
-- Create index "items_ns_created_idx" to table: "items"
CREATE INDEX "items_ns_created_idx" ON "public"."items" ("org_id", "namespace", "created_at");
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
