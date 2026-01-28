create table if not exists foo (
  id text primary key,
  org_id text not null,
  namespace text not null,
  name text not null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create index if not exists foo_org_ns_created_idx
  on foo(org_id, namespace, created_at desc);

create unique index if not exists foo_org_ns_name_uniq
  on foo(org_id, namespace, lower(name));
