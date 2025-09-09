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
  column "embedding" {
    type = sql("vector(384)")
    null = true
    comment = "Text embedding for ANN similarity"
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
  index "items_org_ns_available_item_idx" {
    columns = [
      column.org_id,
      column.namespace,
      column.available,
      column.item_id,
    ]
  }
  index "items_tags_gin_idx" {
    type    = GIN
    columns = [column.tags]
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
  column "source_event_id" {
    type = text
    null = true
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
  index "events_source_uidx" {
    unique  = true
    columns = [ column.org_id, column.namespace, column.source_event_id ]
  }
}

table "event_type_defaults" {
  schema = schema.public

  column "type" {
    type = smallint
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "weight" {
    type = double_precision
    null = false
  }
  column "half_life_days" {
    type = double_precision
    null = true
  }

  primary_key {
    columns = [column.type]
  }

  check "event_type_defaults_weight_positive" {
    expr = "weight > 0"
  }
  check "event_type_defaults_hl_positive" {
    expr = "(half_life_days IS NULL) OR (half_life_days > 0)"
  }
}

table "event_type_config" {
  schema = schema.public

  column "org_id" {
    type = uuid
    null = false
  }
  column "namespace" {
    type = text
    null = false
  }
  column "type" {
    type = smallint
    null = false
  }
  column "name" {
    type = text
    null = true
  }
  column "weight" {
    type = double_precision
    null = false
  }
  column "half_life_days" {
    type = double_precision
    null = true
  }
  column "is_active" {
    type = boolean
    null = false
    default = true
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }

  primary_key {
    columns = [column.org_id, column.namespace, column.type]
  }

  index "event_type_config_org_ns_active_idx" {
    columns = [column.org_id, column.namespace]
    where   = "is_active = true"
  }

  check "event_type_config_weight_positive" {
    expr = "weight > 0"
  }
  check "event_type_config_hl_positive" {
    expr = "(half_life_days IS NULL) OR (half_life_days > 0)"
  }
  check "event_type_config_type_nonneg" {
    expr = "type >= 0"
  }
}
