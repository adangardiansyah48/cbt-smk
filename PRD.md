# PRODUCT REQUIREMENT DOCUMENT (PRD)

## Sistem Ujian CBT Terpadu SMK (UAS/PAS/UKK/UKOM) Berkonsep ANBK & TKA

| Nama Proyek | Sistem CBT Ujian Sekolah & Kejuruan SMK (Konsep ANBK & TKA) |
| --- | --- |
| Spesifikasi Lembaga | Sekolah Menengah Kejuruan (SMK) - Konsentrasi Keahlian & PKL/Magang |
| Cakupan Ujian | UAS/PAS, Asesmen Sumatif, Tes Teori UKK/UKOM, Ujian Sertifikasi Produktif |
| Versi Document | v2.0 (Khusus Struktur SMK Lengkap) |
| Tech Stack Utama | Vue 3 (SPA/PWA), Golang API, Supabase (PostgreSQL), Cloudflare, Render |
| Tanggal Penyusunan | September 2026 |

### 1. Analisis Kebutuhan & Spesifikasi Khusus Sekolah Menengah Kejuruan (SMK)

Penyelenggaraan ujian di Sekolah Menengah Kejuruan (SMK) memiliki kompleksitas struktur yang jauh berbeda dibanding SMA/SMP. Selain ujian teori norma adaptif (UAS/PAS), SMK membutuhkan fleksibilitas pengelompokan mata pelajaran produktif berdasarkan Bidang Keahlian, Program Keahlian, dan Konsentrasi Keahlian (Jurusan).

Sistem CBT ini dirancang khusus untuk memenuhi ekosistem SMK, mendukung ujian Teori Kejuruan (UKK/UKOM/LSP), serta mampu mengelola status siswa yang sedang melaksanakan Praktik Kerja Lapangan (PKL/Magang Industrial) sehingga dapat mengikuti ujian secara remote/daring dari lokasi kerja dengan aman.

| 📌 FITUR UTAMA BERBASIS SPESIFIKASI SMKSistem mengakomodasi Hierarki Akademik SMK: Sekolah -> Program/Konsentrasi Keahlian (TKJ, RPL, TKR, AKL, dll) -> Bengkel/Lab/Kelas -> Rombel. Mendukung soal bergambar/diagram teknik/skema rangkaian, media audio (Listening Bahasa Asing), serta status siswa PKL. |
| --- |

### 2. Spesifikasi Arsitektur Infrastructure & Free-Tier Strategy

| Komponen | Teknologi Terpilih | Optimasi & Batas Free-Tier |
| --- | --- | --- |
| Frontend Client | Vue 3 + Vite + Tailwind CSS (PWA) | Offline-first resilience via Service Worker & LocalStorage (cocok untuk HP siswa PKL di daerah minim sinyal). |
| Static Hosting & CDN | Cloudflare Pages | Unlimited bandwidth. Global Edge Network mempercepat akses gambar diagram teknik & skema soal kejuruan. |
| Backend API | Golang (Fiber Framework) | High-concurrency ringan (RAM < 30MB). Menangani request serentak ratusan siswa saat jam mulai ujian. |
| Backend Hosting | Render (Free Web Service) | Dioptimalkan dengan Cron Keep-Alive /ping tiap 10 menit agar backend tidak tidur (mencegah cold-start). |
| Database & Auth | Supabase (PostgreSQL) | PostgreSQL 500MB DB. Penggunaan RLS Policy & skema JSONB untuk opsi soal fleksibel dan enkripsi JWT Auth. |

### 3. Kebutuhan Fungsional & Modul Aplikasi SMK

### 3.1 Modul Peserta Ujian (Siswa SMK - Mobile First PWA)

• Multi-Device Support: Responsif di Smartphone, Tablet, PC Bengkel/Lab Komputer SMK.

• Otentikasi Siswa & Sesi PKL: Login NISN & Token Ujian. Dukungan mode khusus 'Ujian Remote' untuk siswa yang sedang PKL/Magang.

• Tipe Soal Berstandar ANBK/TKA Kejuruan:

- Pilihan Ganda Biasa (1 Jawaban Benar)

- Pilihan Ganda Kompleks (Pilihan Multi-Jawaban Benar, e.g. Troubleshooting Komponen Engine)

- Menjodohkan (Memasangkan Simbol/Diagram dengan Fungsi Komponen)

- Isian Singkat & Essay/Uraian (Analisis Kasus / Langkah Pembacaan Alat Ukur)

• Media Rich Question Support: Optimal dalam menangani gambar diagram skematik, wiring diagram, flowchart, dan audio listening.

• Auto-Save LocalSync: Jawaban tersimpan di memori lokal HP setiap 1 detik dan di-sync ke server Go tiap 3-5 detik.

• Anti-Cheat Protocol: Peringatan otomatis saat siswa pindah tab/aplikasi, pindah mode layar, serta blokir copy-paste & right-click.

### 3.2 Modul Pengawas, Kepala Bengkel/Kaprog & Guru Produktif

• Manajemen Jurusan & Lab: Pengelompokan soal berdasarkan Konsentrasi Keahlian dan Laboratorium/Bengkel.

• Bank Soal Kejuruan & Import Massal: Import bank soal via Excel/JSON beserta lampiran file media (Gambar Diagram/Audio).

• Jadwal Ujian & Generator Token: Pengaturan jadwal fleksibel, durasi pengerjaan, acak urutan soal & opsi, serta token per mapel/jurusan.

• Real-time Monitoring Dashboard: Memantau keaktifan siswa (Status Online, Jawaban Terisi, Selesai, atau Terputus) per rombel/jurusan.

• Auto-Grading & Rekap Nilai Rapor/UKK: Penilaian otomatis instan untuk PG, PG Kompleks, Menjodohkan. Hasil terkap per jurusan dan siap diexport ke Excel.

### 4. Rancangan Skema Database Terintegrasi (PostgreSQL / Supabase)

Berikut adalah struktur skema database PostgreSQL terlengkap yang mengakomodasi seluruh data master SMK, jurusan, guru pengampu, bank soal kejuruan, hingga rekap jawaban siswa:

| Nama Tabel | Kolom Utama / Foreign Keys | Fungsi & Deskripsi Modul SMK |
| --- | --- | --- |
| school_profiles | id, school_name, npsn, address, headmaster_name, logo_url | Identitas Sekolah SMK, KOP Surat, Berita Acara Ujian & Kartu Peserta. |
| academic_years | id, year_name (e.g. 2026/2027), is_active | Master Tahun Ajaran aktif. |
| semesters | id, academic_year_id, semester_type (ganjil/genap), is_active | Master Semester berjalan. |
| majors | id, code (TKJ/RPL/TKR/AKL), name, department_head | Master Konsentrasi Keahlian / Jurusan SMK. |
| classes | id, major_id, grade_level (10/11/12), class_name | Master Kelas & Rombel (e.g., XII TKJ 1, XI TKR 2). |
| subjects | id, major_id, code, name, pass_grade (KKM), subject_type | Master Mapel (Normatif, Adaptif, Produktif Kejuruan). |
| users | id (auth.users), username, full_name, role (admin/teacher/student) | Master Akun terintegrasi Supabase Auth JWT. |
| teachers | id, user_id, nip_nuptk, is_internal, gender | Master Data Guru / Penguji UKK External. |
| students | id, user_id, nisn, nis, class_id, is_pkl_active | Master Siswa SMK (dilengkapi flag status siswa PKL). |
| teacher_subjects | id, teacher_id, subject_id, class_id, academic_year_id | Mapping Guru Pengampu Mapel per Kelas & Jurusan. |
| questions | id, subject_id, teacher_id, question_type, content, options (JSONB) | Bank Soal Kejuruan (Soal, Diagram, Kunci Jawaban JSONB). |
| exams | id, title, subject_id, semester_id, duration, token, is_pkl_allowed | Sesi Ujian/UKK (dilengkapi izin pengerjaan remote PKL). |
| exam_classes | id, exam_id, class_id | Target Rombel/Kelas yang mengikuti sesi ujian. |
| exam_questions | id, exam_id, question_id, order_no | Pivot daftar soal yang diujikan dalam satu paket. |
| exam_responses | id, exam_id, student_id, question_id, answer (JSONB), is_flagged | Penyimpanan Jawaban Siswa real-time (Upsert JSONB). |
| exam_results | id, exam_id, student_id, total_pg_score, total_essay_score, final_score | Rekap Nilai Akhir Siswa per Ujian & Status Selesai. |

### 5. Rencana Tahapan Pengembangan & Skenario Pengujian (Roadmap)

| Fase | Fokus Aktivitas SMK | Deliverables / Target Output |
| --- | --- | --- |
| Fase 1 | Database SMK & Core Go API | Setup 16 Tabel Supabase, RLS Policy, API Auth NISN, Import Siswa/Guru per Jurusan, Engine Upsert Jawaban. |
| Fase 2 | Vue 3 CBT Mobile Interface | UI Ujian Responsif Mobile/Lab SMK, Support Diagram/Media, Pinia Store, LocalStorage Offline Sync, Anti-Cheat. |
| Fase 3 | Modul Jurusan & Auto-Grading | Dashboard Pengawas/Kaprog, Import Soal Excel + Media Gambar, Token Ujian, Auto-Grading Engine, Rekap Nilai Rapor. |
| Fase 4 | Deployment & Stress Testing | CI/CD Cloudflare Pages & Render, Cron UptimeRobot, Load Test (500+ siswa bersamaan dari Lab Komputer SMK). |
