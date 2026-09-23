-- CBT SMK — PostgreSQL / Supabase schema
-- 16 tabel + indeks + RLS + seed

create extension if not exists "pgcrypto";

-- ============================================================
-- TABLES
-- ============================================================

create table if not exists school_profiles (
  id uuid primary key default gen_random_uuid(),
  school_name text not null,
  npsn text unique,
  address text,
  headmaster_name text,
  logo_url text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists academic_years (
  id uuid primary key default gen_random_uuid(),
  year_name text not null,
  is_active boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists semesters (
  id uuid primary key default gen_random_uuid(),
  academic_year_id uuid not null references academic_years(id) on delete cascade,
  semester_type text not null check (semester_type in ('ganjil', 'genap')),
  is_active boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists majors (
  id uuid primary key default gen_random_uuid(),
  code text not null unique,
  name text not null,
  department_head text,
  created_at timestamptz not null default now()
);

create table if not exists classes (
  id uuid primary key default gen_random_uuid(),
  major_id uuid not null references majors(id) on delete cascade,
  grade_level int not null check (grade_level in (10, 11, 12)),
  class_name text not null,
  created_at timestamptz not null default now(),
  unique (major_id, class_name)
);

create table if not exists subjects (
  id uuid primary key default gen_random_uuid(),
  major_id uuid references majors(id) on delete set null,
  code text not null,
  name text not null,
  pass_grade numeric(5,2) not null default 75,
  subject_type text not null check (subject_type in ('normatif', 'adaptif', 'produktif')),
  created_at timestamptz not null default now()
);

create table if not exists users (
  id uuid primary key default gen_random_uuid(),
  auth_id uuid unique,
  username text unique not null,
  full_name text not null,
  role text not null check (role in ('admin', 'teacher', 'student')),
  password_hash text,
  created_at timestamptz not null default now()
);

create table if not exists teachers (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null unique references users(id) on delete cascade,
  nip_nuptk text,
  is_internal boolean not null default true,
  gender text check (gender in ('L', 'P')),
  created_at timestamptz not null default now()
);

create table if not exists students (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null unique references users(id) on delete cascade,
  nisn text unique not null,
  nis text,
  class_id uuid references classes(id) on delete set null,
  is_pkl_active boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists teacher_subjects (
  id uuid primary key default gen_random_uuid(),
  teacher_id uuid not null references teachers(id) on delete cascade,
  subject_id uuid not null references subjects(id) on delete cascade,
  class_id uuid not null references classes(id) on delete cascade,
  academic_year_id uuid not null references academic_years(id) on delete cascade,
  created_at timestamptz not null default now(),
  unique (teacher_id, subject_id, class_id, academic_year_id)
);

create table if not exists questions (
  id uuid primary key default gen_random_uuid(),
  subject_id uuid not null references subjects(id) on delete cascade,
  teacher_id uuid references teachers(id) on delete set null,
  question_type text not null check (question_type in ('pg', 'pg_complex', 'matching', 'short', 'essay')),
  content text not null,
  media_url text,
  media_type text check (media_type is null or media_type in ('image', 'audio')),
  options jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists exams (
  id uuid primary key default gen_random_uuid(),
  title text not null,
  subject_id uuid not null references subjects(id) on delete cascade,
  semester_id uuid references semesters(id) on delete set null,
  duration int not null default 90,
  token text not null,
  is_pkl_allowed boolean not null default false,
  shuffle_questions boolean not null default true,
  shuffle_options boolean not null default true,
  start_at timestamptz,
  end_at timestamptz,
  is_active boolean not null default false,
  created_by uuid references users(id) on delete set null,
  created_at timestamptz not null default now()
);

create table if not exists exam_classes (
  id uuid primary key default gen_random_uuid(),
  exam_id uuid not null references exams(id) on delete cascade,
  class_id uuid not null references classes(id) on delete cascade,
  unique (exam_id, class_id)
);

create table if not exists exam_questions (
  id uuid primary key default gen_random_uuid(),
  exam_id uuid not null references exams(id) on delete cascade,
  question_id uuid not null references questions(id) on delete cascade,
  order_no int not null default 0,
  unique (exam_id, question_id)
);

create table if not exists exam_responses (
  id uuid primary key default gen_random_uuid(),
  exam_id uuid not null references exams(id) on delete cascade,
  student_id uuid not null references students(id) on delete cascade,
  question_id uuid not null references questions(id) on delete cascade,
  answer jsonb not null default 'null'::jsonb,
  is_flagged boolean not null default false,
  updated_at timestamptz not null default now(),
  unique (exam_id, student_id, question_id)
);

create table if not exists exam_results (
  id uuid primary key default gen_random_uuid(),
  exam_id uuid not null references exams(id) on delete cascade,
  student_id uuid not null references students(id) on delete cascade,
  total_pg_score numeric(6,2) not null default 0,
  total_essay_score numeric(6,2) not null default 0,
  final_score numeric(6,2) not null default 0,
  status text not null default 'in_progress' check (status in ('in_progress', 'submitted', 'graded')),
  started_at timestamptz not null default now(),
  submitted_at timestamptz,
  last_seen_at timestamptz not null default now(),
  cheat_warnings int not null default 0,
  unique (exam_id, student_id)
);

-- ============================================================
-- INDEXES
-- ============================================================

create index if not exists idx_semesters_year on semesters (academic_year_id);
create index if not exists idx_classes_major on classes (major_id);
create index if not exists idx_subjects_major on subjects (major_id);
create index if not exists idx_students_class on students (class_id);
create index if not exists idx_students_nisn on students (nisn);
create index if not exists idx_teachers_user on teachers (user_id);
create index if not exists idx_questions_subject on questions (subject_id);
create index if not exists idx_exams_token on exams (token);
create index if not exists idx_exams_subject on exams (subject_id);
create index if not exists idx_exam_classes_exam on exam_classes (exam_id);
create index if not exists idx_exam_questions_exam on exam_questions (exam_id);
create index if not exists idx_exam_responses_lookup on exam_responses (exam_id, student_id);
create index if not exists idx_exam_results_exam on exam_results (exam_id);
create index if not exists idx_exam_results_status on exam_results (exam_id, status);

-- ============================================================
-- UPDATED_AT TRIGGER
-- ============================================================

create or replace function set_updated_at()
returns trigger language plpgsql as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

drop trigger if exists trg_school_profiles_updated on school_profiles;
create trigger trg_school_profiles_updated
  before update on school_profiles
  for each row execute function set_updated_at();

drop trigger if exists trg_questions_updated on questions;
create trigger trg_questions_updated
  before update on questions
  for each row execute function set_updated_at();

drop trigger if exists trg_exam_responses_updated on exam_responses;
create trigger trg_exam_responses_updated
  before update on exam_responses
  for each row execute function set_updated_at();

-- ============================================================
-- RLS
-- ============================================================

alter table school_profiles enable row level security;
alter table academic_years enable row level security;
alter table semesters enable row level security;
alter table majors enable row level security;
alter table classes enable row level security;
alter table subjects enable row level security;
alter table users enable row level security;
alter table teachers enable row level security;
alter table students enable row level security;
alter table teacher_subjects enable row level security;
alter table questions enable row level security;
alter table exams enable row level security;
alter table exam_classes enable row level security;
alter table exam_questions enable row level security;
alter table exam_responses enable row level security;
alter table exam_results enable row level security;

-- Authenticated staff can read master data
do $$
declare
  t text;
begin
  foreach t in array array[
    'school_profiles','academic_years','semesters','majors','classes','subjects',
    'users','teachers','students','teacher_subjects','questions','exams',
    'exam_classes','exam_questions','exam_responses','exam_results'
  ]
  loop
    execute format('drop policy if exists staff_read on %I', t);
    execute format(
      'create policy staff_read on %I for select to authenticated using (true)',
      t
    );
  end loop;
end;
$$;

-- Staff write (admin/teacher via JWT claim role, fallback authenticated)
drop policy if exists staff_write_questions on questions;
create policy staff_write_questions on questions
  for all to authenticated
  using (true)
  with check (true);

drop policy if exists staff_write_exams on exams;
create policy staff_write_exams on exams
  for all to authenticated
  using (true)
  with check (true);

drop policy if exists staff_write_exam_classes on exam_classes;
create policy staff_write_exam_classes on exam_classes
  for all to authenticated
  using (true)
  with check (true);

drop policy if exists staff_write_exam_questions on exam_questions;
create policy staff_write_exam_questions on exam_questions
  for all to authenticated
  using (true)
  with check (true);

drop policy if exists staff_write_results on exam_results;
create policy staff_write_results on exam_results
  for all to authenticated
  using (true)
  with check (true);

drop policy if exists staff_write_responses on exam_responses;
create policy staff_write_responses on exam_responses
  for all to authenticated
  using (true)
  with check (true);

drop policy if exists staff_write_master on majors;
create policy staff_write_master on majors
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_classes on classes;
create policy staff_write_classes on classes
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_subjects on subjects;
create policy staff_write_subjects on subjects
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_students on students;
create policy staff_write_students on students
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_teachers on teachers;
create policy staff_write_teachers on teachers
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_users on users;
create policy staff_write_users on users
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_ts on teacher_subjects;
create policy staff_write_ts on teacher_subjects
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_school on school_profiles;
create policy staff_write_school on school_profiles
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_years on academic_years;
create policy staff_write_years on academic_years
  for all to authenticated using (true) with check (true);

drop policy if exists staff_write_semesters on semesters;
create policy staff_write_semesters on semesters
  for all to authenticated using (true) with check (true);

-- Anon read for public school profile (kartu peserta / login splash)
drop policy if exists anon_read_school on school_profiles;
create policy anon_read_school on school_profiles
  for select to anon using (true);

-- ============================================================
-- SEED
-- ============================================================

insert into school_profiles (school_name, npsn, address, headmaster_name)
select 'SMK Negeri Contoh', '12345678', 'Jl. Pendidikan No. 1', 'Drs. Kepala Sekolah'
where not exists (select 1 from school_profiles);

insert into academic_years (year_name, is_active)
select '2026/2027', true
where not exists (select 1 from academic_years where year_name = '2026/2027');

insert into semesters (academic_year_id, semester_type, is_active)
select ay.id, 'ganjil', true
from academic_years ay
where ay.year_name = '2026/2027'
  and not exists (
    select 1 from semesters s
    where s.academic_year_id = ay.id and s.semester_type = 'ganjil'
  );

insert into majors (code, name, department_head)
select v.code, v.name, v.head
from (values
  ('RPL', 'Rekayasa Perangkat Lunak', 'Kaprog RPL'),
  ('TKJ', 'Teknik Komputer dan Jaringan', 'Kaprog TKJ'),
  ('TKR', 'Teknik Kendaraan Ringan', 'Kaprog TKR'),
  ('AKL', 'Akuntansi dan Keuangan Lembaga', 'Kaprog AKL')
) as v(code, name, head)
where not exists (select 1 from majors m where m.code = v.code);

insert into classes (major_id, grade_level, class_name)
select m.id, 12, 'XII ' || m.code || ' 1'
from majors m
where not exists (
  select 1 from classes c where c.major_id = m.id and c.class_name = 'XII ' || m.code || ' 1'
);

insert into subjects (major_id, code, name, pass_grade, subject_type)
select m.id, 'PWL', 'Pemrograman Web Lanjut', 75, 'produktif'
from majors m where m.code = 'RPL'
  and not exists (select 1 from subjects s where s.code = 'PWL');

insert into subjects (major_id, code, name, pass_grade, subject_type)
select m.id, 'TKJ-JARKOM', 'Jaringan Komputer', 75, 'produktif'
from majors m where m.code = 'TKJ'
  and not exists (select 1 from subjects s where s.code = 'TKJ-JARKOM');

insert into subjects (major_id, code, name, pass_grade, subject_type)
select null, 'BIN', 'Bahasa Indonesia', 75, 'normatif'
where not exists (select 1 from subjects s where s.code = 'BIN');

-- migration add password_hash if tabel lama
alter table users add column if not exists password_hash text;

-- admin default: username admin / password admin123 (hash sesuai HashPassword: sha256("cbt-smk:admin123"))
-- sha256hex("cbt-smk:admin123") = hitung oleh backend HashPassword; placeholder diisi saat migrate via trigger tugas seeder handler
do $$
declare v text;
begin
  select encode(digest('cbt-smk:admin123', 'sha256'), 'hex') into v;
  if not exists (select 1 from users u where u.username = 'admin') then
    insert into users (username, full_name, role, password_hash)
    values ('admin', 'Administrator CBT', 'admin', v);
  end if;
end
$$;
