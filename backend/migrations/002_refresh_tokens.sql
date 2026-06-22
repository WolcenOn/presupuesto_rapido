create table if not exists refresh_tokens (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  token_hash bytea not null unique,
  expires_at timestamptz not null,
  revoked_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists idx_refresh_tokens_user_active on refresh_tokens(user_id, expires_at desc) where revoked_at is null;

create table if not exists companies (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  tax_id text,
  email text,
  phone text,
  address text,
  city text,
  postal text,
  province text,
  country text not null default 'ES',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

alter table users add column if not exists company_id uuid references companies(id);
alter table price_items add column if not exists company_id uuid references companies(id);
alter table documents add column if not exists company_id uuid references companies(id);
alter table employee_mail_settings add column if not exists company_id uuid references companies(id);
alter table audit_logs add column if not exists company_id uuid references companies(id);
alter table document_email_logs add column if not exists company_id uuid references companies(id);

create index if not exists idx_users_company_role on users(company_id, role);
create index if not exists idx_price_items_company_name on price_items(company_id, name);
create index if not exists idx_documents_company_created on documents(company_id, created_at desc);
create index if not exists idx_documents_company_employee_created on documents(company_id, employee_id, created_at desc);
create index if not exists idx_email_logs_company_status on document_email_logs(company_id, status, created_at);
