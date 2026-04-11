create extension if not exists pgcrypto;

create table users (
  id uuid primary key default gen_random_uuid(),
  email varchar(255) not null unique,
  password_hash varchar(255) not null,
  display_name varchar(100) not null,
  role varchar(20) not null default 'editor',
  is_active boolean not null default true,
  last_login_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_users_role
    check (role in ('admin', 'editor', 'viewer'))
);

create table categories (
  id uuid primary key default gen_random_uuid(),
  slug varchar(100) not null unique,
  name varchar(100) not null unique,
  description text,
  icon varchar(50),
  sort_order integer not null default 0,
  is_visible boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table tags (
  id uuid primary key default gen_random_uuid(),
  slug varchar(100) not null unique,
  name varchar(100) not null unique,
  created_at timestamptz not null default now()
);

create table tools (
  id uuid primary key default gen_random_uuid(),
  slug varchar(150) not null unique,
  name varchar(150) not null unique,
  short_description varchar(280) not null,
  long_description text not null,
  category_id uuid not null references categories(id),
  difficulty varchar(20) not null,
  icon varchar(50),
  featured boolean not null default false,
  status varchar(20) not null default 'draft',
  website_url text not null,
  github_url text,
  view_count integer not null default 0,
  favorite_count integer not null default 0,
  published_at timestamptz,
  created_by uuid references users(id),
  updated_by uuid references users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_tools_difficulty
    check (difficulty in ('beginner', 'intermediate', 'advanced', 'expert')),
  constraint chk_tools_status
    check (status in ('draft', 'published', 'archived'))
);

create table tool_tags (
  tool_id uuid not null references tools(id) on delete cascade,
  tag_id uuid not null references tags(id) on delete cascade,
  primary key (tool_id, tag_id)
);

create table tool_submissions (
  id uuid primary key default gen_random_uuid(),
  submitted_by uuid references users(id),
  tool_id uuid references tools(id),
  submitter_email varchar(255),
  payload jsonb not null,
  status varchar(20) not null default 'pending',
  reviewer_id uuid references users(id),
  review_note text,
  created_at timestamptz not null default now(),
  reviewed_at timestamptz,
  constraint chk_tool_submissions_status
    check (status in ('pending', 'approved', 'rejected'))
);

create table audit_logs (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references users(id),
  action varchar(50) not null,
  resource_type varchar(50) not null,
  resource_id uuid,
  before_data jsonb,
  after_data jsonb,
  created_at timestamptz not null default now()
);

create index idx_categories_sort_order on categories(sort_order);
create index idx_tools_category_id on tools(category_id);
create index idx_tools_status on tools(status);
create index idx_tools_featured on tools(featured);
create index idx_tools_published_at on tools(published_at desc);
create index idx_tools_category_status on tools(category_id, status);
create index idx_tool_submissions_status on tool_submissions(status);
create index idx_audit_logs_user_id on audit_logs(user_id);
create index idx_audit_logs_resource on audit_logs(resource_type, resource_id);
