create table if not exists template_versions (
  id uuid primary key default gen_random_uuid(),
  template_id uuid not null references templates(id) on delete restrict,
  version int not null,
  status text not null check (status in ('draft','published')),
  content jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  unique (template_id, version)
);

-- Only one published version per template
create unique index if not exists ux_template_versions_one_published
  on template_versions(template_id)
  where status = 'published';
