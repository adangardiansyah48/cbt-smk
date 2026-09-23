# CBT SMK — Sistem Ujian Terpadu SMK (UAS/PAS/UKK/UKOM) Konsep ANBK & TKA

Stack: **Vue 3 + Vite + Tailwind (PWA)** · **Golang Fiber API** · **Supabase PostgreSQL** · Cloudflare Pages · Render

> Dibuat dari PRD `PRD_Sistem_Ujian_CBT_SMK_ANBK_TKA.docx` → `PRD.md`.

## Struktur

```
D:\cbt\
├── PRD.md
├── database\schema.sql     # 16 tabel + RLS + seed (majors, classes, subjects, admin)
├── backend\                # Go Fiber API
│   ├── .env                # DATABASE_URL pooler + SUPABASE_*
│   ├── config/  handlers/  middleware/  models/
│   ├── main.go
│   ├── Dockerfile          # multi-stage build untuk Render
│   ├── render.yaml         # konfigurasi Render Web Service
│   └── .dockerignore
├── frontend\               # Vue 3 + Vite + Tailwind + Pinia + vue-router
│   ├── .env
│   ├── .env.production.example
│   ├── cloudflare-pages.toml
│   ├── src/pages/Login.vue  # NISN+token / admin
│   ├── src/pages/Exam.vue   # PG, PG kompleks, matching, essay + auto-save + anti-cheat
│   └── src/pages/Admin.vue  # Majors/classes/subjects/students/teachers/questions/exams/monitor
├── .github/workflows/deploy.yml
└── README.md
```

## Database (Supabase)

Sudah di-apply ke proyek Supabase:

- URL: `https://znghhzfkuxqdokokqdwr.supabase.co`
- Pooler (yang dipakai backend, karena `db.*` hanya IPv6 di lokal): `aws-0-ap-southeast-1.pooler.supabase.com:6543`
- 16 tabel: `school_profiles`, `academic_years`, `semesters`, `majors`, `classes`, `subjects`, `users`, `teachers`, `students`, `teacher_subjects`, `questions`, `exams`, `exam_classes`, `exam_questions`, `exam_responses`, `exam_results`
- Seed: 4 jurusan (RPL/TKJ/TKR/AKL), 4 kelas XII, 3 mapel (PWL, JARKOM, BIN), admin `admin / admin123`.

Jika perlu re-apply:

```powershell
node --input-type=module -e "
import pg from 'pg'; import fs from 'fs';
const sql=fs.readFileSync('D:/cbt/database/schema.sql','utf8');
const c=new pg.Client({connectionString:'postgresql://postgres.znghhzfkuxqdokokqdwr:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres', ssl:{rejectUnauthorized:false}});
await c.connect(); await c.query(sql); await c.end();
"
```

## Backend

```powershell
$env:Path="C:\Program Files\Go\bin;"+$env:Path
cd D:\cbt\backend
go mod tidy
go build -o ./tmp/cbt-api.exe .
./tmp/cbt-api.exe   # :8080
```

Endpoints utama:
- `POST /api/auth/login` — `{nisn, token}` → JWT student + row `exam_results`
- `POST /api/auth/staff` — `{username, password}` → JWT admin/teacher
- `GET /api/exam/paper` — daftar soal (options tanpa `correct`, acak jika diaktifkan)
- `POST /api/exam/response` — upsert jawaban JSONB
- `POST /api/exam/heartbeat` — keep-alive + cheat counter
- `POST /api/exam/submit` — auto-grading PG/pg_complex/matching
- `GET/POST /api/majors /classes /subjects /students /teachers /questions /exams`
- `GET /api/exams/:id/results` `…/monitor` `POST …/grade`
- `GET /ping` — health check untuk Render cron

## Frontend

```powershell
cd D:\cbt\frontend
npm install
npm run dev    # :5173, proxy /api → :8080
npm run build  # dist/
```

Alur pakai:
1. Admin login `admin / admin123` → **Admin** → buat soal, buat ujian (token auto 6 char), salin token.
2. Siswa login tab **Siswa** → NISN + token ujian → kerjakan (PG, kompleks, matching, essay), jawaban auto-save tiap 1 detik ke LocalStorage + 3–5 detik ke server, pindah tab = warning + `cheat_warnings`.
3. Admin **Monitor** → live status (online via `last_seen_at`), **Hasil** → rekap nilai → **Export CSV**.
4. Tombol **Kumpulkan** → grading auto untuk PG, essay menunggu input manual (`total_essay_score`).

## Akun Demo (sudah dibuat)

| Role | Kredensial |
|------|------------|
| Admin | `admin / admin123` |
| Siswa | NISN `0050000001` (kelas XII RPL 1) — token ujian `K9CFN4` (UKK Teori RPL Demo, 4 soal) |

Tambah siswa/guru/soal/ujian baru lewat Admin UI atau via `curl` dengan header `Authorization: Bearer <staff_token>`.

## Deploy (sesuai PRD Fase 4)

### 1. Frontend → Cloudflare Pages

**Opsi A: Via Dashboard (recommended)**
1. Push repo ke GitHub.
2. Buka Cloudflare Pages → **Create a project** → **Connect to Git**.
3. Pilih repo, **Framework preset**: `Vite`.
4. Build command: `npm run build` | Output directory: `dist` | Root directory: `frontend`.
5. **Environment variables** (Add variable):
   - `VITE_API_URL` = URL backend Render (contoh: `https://cbt-api.onrender.com`)
   - `VITE_SUPABASE_URL` = `https://znghhzfkuxqdokokqdwr.supabase.co`
   - `VITE_SUPABASE_ANON_KEY` = `<anon key dari Supabase>`
6. **Save and Deploy**.

**Opsi B: Wrangler CLI**
```bash
cd frontend
npm install -g wrangler
wrangler pages deploy dist --project-name=cbt-smk
```

### 2. Backend → Render (Free Web Service)

**Opsi A: Via Dashboard (recommended)**
1. Push repo ke GitHub.
2. Buka Render Dashboard → **New +** → **Web Service** → **Docker** (atau **Go**).
3. Connect repository.
4. **Environment**: `Docker` (auto-detect `backend/Dockerfile`).
5. **Environment Variables** (Add secret):
   - `DATABASE_URL` = `postgresql://postgres.PROJECT_REF:PASSWORD@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres?sslmode=require`
   - `SUPABASE_URL` = `https://PROJECT_REF.supabase.co`
   - `SUPABASE_ANON_KEY` = `<anon key>`
   - `SUPABASE_SERVICE_ROLE_KEY` = `<service_role key>`
   - `JWT_SECRET` = generate random 32+ chars (contoh: `openssl rand -hex 32`)
   - `FRONTEND_ORIGIN` = URL Cloudflare Pages (contoh: `https://cbt-smk.pages.dev`)
   - `PORT` = `8080`
6. **Health Check Path**: `/ping`
7. **Deploy**.

**Opsi B: render.yaml (Infrastructure as Code)**
Render mendukung `render.yaml` di root repo. File sudah disediakan di `backend/render.yaml`.

### 3. Cron Keep-Alive (UptimeRobot / Render Cron)

Render Free tier mematikan service setelah 15 menit idle. Tambahkan cron external:

1. Daftar di [UptimeRobot](https://uptimerobot.com/) (free).
2. Buat monitor **HTTP(s)** → URL: `https://<render-service>.onrender.com/ping`
3. Interval: **10 menit**.
4. Save.

Atau gunakan **Render Cron Job** (paid add-on) atau GitHub Actions scheduled workflow.

### 4. CI/CD GitHub Actions (Opsional)

File `.github/workflows/deploy.yml` sudah tersedia:

```yaml
name: Deploy CBT SMK
on:
  push:
    branches: [main]

jobs:
  deploy-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 24 }
      - working-directory: frontend
        run: |
          npm ci
          npm run build
      - uses: cloudflare/pages-action@v1
        with:
          apiToken: ${{ secrets.CF_API_TOKEN }}
          accountId: ${{ secrets.CF_ACCOUNT_ID }}
          projectName: cbt-smk
          directory: frontend/dist
          gitHubToken: ${{ secrets.GITHUB_TOKEN }}

  deploy-backend:
    needs: deploy-frontend
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - working-directory: backend
        run: |
          go mod download
          go build -o cbt-api .
          # Render auto-deploys dari GitHub, cukup build & test
```

**Secrets yang dibutuhkan** (GitHub repo Settings → Secrets):
- `CF_API_TOKEN` — Cloudflare API token (Pages:Edit)
- `CF_ACCOUNT_ID` — Cloudflare Account ID

## Quick Test Lokal

```powershell
# Terminal 1 - Backend
$env:Path="C:\Program Files\Go\bin;"+$env:Path
cd D:\cbt\backend
go run .

# Terminal 2 - Frontend
cd D:\cbt\frontend
npm run dev

# Buka http://localhost:5173
# Admin: admin / admin123
# Siswa: NISN 0050000001 + Token K9CFN4
```

## Supabase RLS & Auth

- Supabase Auth **tidak dipakai** langsung (custom JWT via Fiber).
- RLS aktif untuk semua tabel, policy `authenticated` = baca/tulis via JWT internal.
- Service role dipakai backend untuk operasi admin (import soal, grading, monitor).

---

**Roadmap PRD:**
- Fase 1 ✅ Database & Core Go API
- Fase 2 ✅ Vue 3 CBT Mobile Interface
- Fase 3 ✅ Modul Jurusan & Auto-Grading
- Fase 4 🚧 Deployment & Stress Testing (Cloudflare Pages + Render + Cron)