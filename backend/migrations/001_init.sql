create extension if not exists pgcrypto;

create table if not exists users (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  email text not null unique,
  password_hash text not null,
  role text not null check (role in ('boss', 'employee')),
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists employee_mail_settings (
  user_id uuid primary key references users(id) on delete cascade,
  smtp_host text not null,
  smtp_port integer not null default 587,
  smtp_username text not null,
  smtp_password_encrypted bytea not null,
  from_email text not null,
  updated_at timestamptz not null default now()
);

create table if not exists price_items (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  base_price numeric(12,2) not null check (base_price >= 0),
  iva_rate numeric(5,2) not null default 21 check (iva_rate >= 0),
  active boolean not null default true,
  updated_by uuid references users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists documents (
  id uuid primary key default gen_random_uuid(),
  ref text not null,
  type text not null check (type in ('presupuesto', 'albaran', 'factura')),
  employee_id uuid not null references users(id),
  client_name text not null,
  client_cif text not null default '',
  client_phone text not null default '',
  client_address text not null default '',
  work_order text,
  payment_method text,
  base_amount numeric(12,2) not null default 0,
  iva_amount numeric(12,2) not null default 0,
  total_amount numeric(12,2) not null default 0,
  document_json jsonb not null default '{}'::jsonb,
  pdf_path text,
  sent_to_boss_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique(employee_id, ref)
);

create index if not exists idx_documents_employee_created on documents(employee_id, created_at desc);
create index if not exists idx_documents_type_created on documents(type, created_at desc);

create table if not exists audit_logs (
  id bigserial primary key,
  user_id uuid references users(id),
  action text not null,
  entity_type text not null,
  entity_id text,
  ip inet,
  user_agent text,
  created_at timestamptz not null default now()
);

create table if not exists document_email_logs (
  id bigserial primary key,
  document_id uuid not null references documents(id) on delete cascade,
  recipient text not null,
  status text not null check (status in ('queued', 'sent', 'failed')),
  error_message text,
  created_at timestamptz not null default now()
);
