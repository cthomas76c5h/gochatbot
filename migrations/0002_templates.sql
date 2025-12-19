create table if not exists templates (
  id uuid primary key default gen_random_uuid(),
  tenant_id uuid not null references tenants(id) on delete restrict,
  name text not null,
  slug text not null,
  created_at timestamptz not null default now(),
  unique (tenant_id, slug)
);
