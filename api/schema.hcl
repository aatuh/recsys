schema "public" {}

table "organizations" {
  schema = schema.public
  column "org_id" {
    type = uuid
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.org_id]
  }
}

table "namespaces" {
  schema = schema.public
  column "id" {
    type = uuid
    null = false
  }
  column "org_id" {
    type = uuid
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  index "namespaces_org_id_name_uq" {
    unique = true
    columns = [column.org_id, column.name]
  }
  foreign_key "namespaces_org_fkey" {
    columns = [column.org_id]
    ref_columns = [table.organizations.column.org_id]
    on_delete = CASCADE
  }
}

table "items" {
  schema = schema.public
  column "org_id" {
    type = uuid
    null = false
  }
  column "namespace" {
    type = text
    null = false
  }
  column "item_id" {
    type = text
    null = false
  }
  column "available" {
    type = boolean
    null = false
    default = true
  }
  column "price" {
    type = numeric(12,2)
    null = true
  }
  column "tags" {
    type = sql("text[]")
    null = false
    default = sql("'{}'::text[]")
  }
  column "vector" {
    type = sql("real[]")
    null = true
  }
  column "props" {
    type = jsonb
    null = false
    default = sql("'{}'::jsonb")
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.org_id, column.namespace, column.item_id]
  }
  index "items_ns_created_idx" {
    columns = [column.org_id, column.namespace, column.created_at]
  }
}

table "users" {
  schema = schema.public
  column "org_id" {
    type = uuid
    null = false
  }
  column "namespace" {
    type = text
    null = false
  }
  column "user_id" {
    type = text
    null = false
  }
  column "traits" {
    type = jsonb
    null = false
    default = sql("'{}'::jsonb")
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.org_id, column.namespace, column.user_id]
  }
  index "users_ns_created_idx" {
    columns = [column.org_id, column.namespace, column.created_at]
  }
}

table "events" {
  schema = schema.public
  column "org_id" {
    type = uuid
    null = false
  }
  column "namespace" {
    type = text
    null = false
  }
  column "user_id" {
    type = text
    null = false
  }
  column "item_id" {
    type = text
    null = false
  }
  column "type" {
    type = smallint
    null = false
  }
  column "value" {
    type = float8
    null = false
    default = 1
  }
  column "ts" {
    type = timestamptz
    null = false
  }
  column "meta" {
    type = jsonb
    null = false
    default = sql("'{}'::jsonb")
  }
  index "events_org_ns_ts_idx" {
    columns = [column.org_id, column.namespace, column.ts]
  }
  index "events_org_ns_user_ts_idx" {
    columns = [column.org_id, column.namespace, column.user_id, column.ts]
  }
  index "events_org_ns_item_ts_idx" {
    columns = [column.org_id, column.namespace, column.item_id, column.ts]
  }
}