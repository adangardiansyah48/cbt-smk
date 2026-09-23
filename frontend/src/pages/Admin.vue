<script setup lang="ts">
import { ref, onMounted, computed } from "vue"
import { apiFetch } from "../lib/api"

const tab = ref("overview")
const majors = ref<any[]>([])
const classes = ref<any[]>([])
const subjects = ref<any[]>([])
const students = ref<any[]>([])
const teachers = ref<any[]>([])
const questions = ref<any[]>([])
const exams = ref<any[]>([])
const results = ref<any[]>([])
const monitor = ref<any[]>([])
const selectedExam = ref("")
const msg = ref(""); const err = ref("")

// forms
const majorForm = ref({ code:"", name:"", department_head:"" })
const classForm = ref({ major_id:"", grade_level:12, class_name:"" })
const subjectForm = ref({ major_id:"", code:"", name:"", pass_grade:75, subject_type:"produktif" })
const studentForm = ref({ username:"", full_name:"", nisn:"", nis:"", class_id:"", is_pkl_active:false })
const teacherForm = ref({ username:"", full_name:"", password:"", nip_nuptk:"", is_internal:true, gender:"L" })
const questionForm = ref({ subject_id:"", question_type:"pg", content:"", media_url:"", media_type:"", optionsText:'[\n  {"id":"A","text":"Opsi A","correct":true},\n  {"id":"B","text":"Opsi B","correct":false}\n]' })
const examForm = ref({ title:"", subject_id:"", duration:90, is_pkl_allowed:false, shuffle_questions:true, shuffle_options:true, is_active:true, class_ids:[] as string[], question_ids:[] as string[] })
const importJson = ref("")

async function loadAll() {
  try {
    const [mj, cl, sj, st, te, qu, ex] = await Promise.all([
      apiFetch("/api/majors"), apiFetch("/api/classes"), apiFetch("/api/subjects"),
      apiFetch("/api/students"), apiFetch("/api/teachers"),
      apiFetch("/api/questions"), apiFetch("/api/exams"),
    ])
    majors.value=mj; classes.value=cl; subjects.value=sj; students.value=st; teachers.value=te; questions.value=qu; exams.value=ex
  } catch (e:any) { err.value=e.message }
}
onMounted(loadAll)

async function createMajor(){ err.value=""; try{ await apiFetch("/api/majors",{method:"POST",body:JSON.stringify(majorForm.value)}); msg.value="Jurusan dibuat"; majorForm.value={code:"",name:"",department_head:""}; loadAll()}catch(e:any){err.value=e.message} }
async function createClass(){ err.value=""; try{ await apiFetch("/api/classes",{method:"POST",body:JSON.stringify(classForm.value)}); msg.value="Kelas dibuat"; classForm.value={major_id:"",grade_level:12,class_name:""}; loadAll()}catch(e:any){err.value=e.message} }
async function createSubject(){ err.value=""; try{ await apiFetch("/api/subjects",{method:"POST",body:JSON.stringify(subjectForm.value)}); msg.value="Mapel dibuat"; subjectForm.value={major_id:"",code:"",name:"",pass_grade:75,subject_type:"produktif"}; loadAll()}catch(e:any){err.value=e.message} }
async function createStudent(){ err.value=""; try{ await apiFetch("/api/students",{method:"POST",body:JSON.stringify(studentForm.value)}); msg.value="Siswa dibuat"; studentForm.value={username:"",full_name:"",nisn:"",nis:"",class_id:"",is_pkl_active:false}; loadAll()}catch(e:any){err.value=e.message} }
async function createTeacher(){ err.value=""; try{ await apiFetch("/api/teachers",{method:"POST",body:JSON.stringify(teacherForm.value)}); msg.value="Guru dibuat"; teacherForm.value={username:"",full_name:"",password:"",nip_nuptk:"",is_internal:true,gender:"L"}; loadAll()}catch(e:any){err.value=e.message} }
async function createQuestion(){
  err.value=""
  try {
    const opts = JSON.parse(questionForm.value.optionsText)
    const payload:any={ subject_id:questionForm.value.subject_id, question_type:questionForm.value.question_type, content:questionForm.value.content, options:opts }
    if(questionForm.value.media_url) { payload.media_url=questionForm.value.media_url; payload.media_type=questionForm.value.media_type||"image" }
    await apiFetch("/api/questions",{method:"POST",body:JSON.stringify(payload)})
    msg.value="Soal dibuat"; questionForm.value.content=""; questionForm.value.optionsText='[]'; loadAll()
  } catch(e:any){ err.value=e.message }
}
async function importQuestions(){
  err.value=""
  if(!importJson.value.trim()) return
  try{
    const arr = JSON.parse(importJson.value)
    const sid = (document.getElementById("import-subject") as HTMLSelectElement)?.value || subjects.value[0]?.id
    if(!sid) { err.value="Pilih mapel dulu"; return }
    const res = await apiFetch(`/api/questions/import?subject_id=${sid}`,{method:"POST",body:JSON.stringify(arr)})
    msg.value=`Import ${res.imported} soal`; importJson.value=""; loadAll()
  }catch(e:any){err.value=e.message}
}
async function createExam(){
  err.value=""
  try{
    const res = await apiFetch("/api/exams",{method:"POST",body:JSON.stringify(examForm.value)})
    msg.value=`Ujian dibuat · Token: ${res.token}`; examForm.value.title=""; loadAll()
  }catch(e:any){err.value=e.message}
}
async function toggleExam(id:string){ await apiFetch(`/api/exams/${id}/toggle`,{method:"POST"}); loadAll() }
async function viewResults(id:string){
  selectedExam.value=id; try{ results.value=await apiFetch(`/api/exams/${id}/results`)}catch(e:any){err.value=e.message}
}
async function viewMonitor(id:string){
  selectedExam.value=id; try{ monitor.value=await apiFetch(`/api/exams/${id}/monitor`)}catch(e:any){err.value=e.message}
}
async function gradeExam(id:string){
  await apiFetch(`/api/exams/${id}/grade`,{method:"POST"}); msg.value="Grading selesai"; viewResults(id)
}
function exportCsv(){
  const rows = results.value
  if(!rows.length) return
  const header = Object.keys(rows[0]).join(",")
  const body = rows.map(r=> Object.values(r).map(v=> `"${String(v??'').replaceAll('"','""')}"`).join(",")).join("\n")
  const blob = new Blob([header+"\n"+body],{type:"text/csv"})
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a"); a.href=url; a.download=`rekap-${selectedExam.value}.csv`; a.click(); URL.revokeObjectURL(url)
}

const filteredQuestions = computed(()=> questions.value)
</script>

<template>
  <div class="mx-auto max-w-6xl p-4">
    <div class="mb-4 flex flex-wrap gap-1 rounded-xl border bg-white p-1 text-sm">
      <button v-for="t in ['overview','majors','classes','subjects','students','teachers','questions','exams','monitor']" :key="t" @click="tab=t" :class="tab===t?'bg-slate-900 text-white':'text-slate-600 hover:bg-slate-100'" class="rounded-lg px-3 py-1.5 capitalize">{{ t }}</button>
    </div>

    <div v-if="msg" class="mb-3 rounded bg-teal-50 p-2 text-sm text-teal-700">{{ msg }}</div>
    <div v-if="err" class="mb-3 rounded bg-red-50 p-2 text-sm text-red-700">{{ err }}</div>

    <!-- OVERVIEW -->
    <div v-if="tab==='overview'" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-xl border bg-white p-4"><div class="text-sm text-slate-500">Jurusan</div><div class="text-2xl font-bold">{{ majors.length }}</div></div>
      <div class="rounded-xl border bg-white p-4"><div class="text-sm text-slate-500">Kelas</div><div class="text-2xl font-bold">{{ classes.length }}</div></div>
      <div class="rounded-xl border bg-white p-4"><div class="text-sm text-slate-500">Siswa</div><div class="text-2xl font-bold">{{ students.length }}</div></div>
      <div class="rounded-xl border bg-white p-4"><div class="text-sm text-slate-500">Ujian</div><div class="text-2xl font-bold">{{ exams.length }}</div></div>
      <div class="rounded-xl border bg-white p-4"><div class="text-sm text-slate-500">Soal</div><div class="text-2xl font-bold">{{ questions.length }}</div></div>
      <div class="rounded-xl border bg-white p-4"><div class="text-sm text-slate-500">Guru</div><div class="text-2xl font-bold">{{ teachers.length }}</div></div>
      <div class="rounded-xl border bg-white p-4"><div class="text-sm text-slate-500">Mapel</div><div class="text-2xl font-bold">{{ subjects.length }}</div></div>
      <div class="rounded-xl border bg-white p-4 text-xs text-slate-500">Gunakan tab untuk kelola data · Buat ujian → bagikan token 6 karakter ke siswa.</div>
    </div>

    <!-- MAJORS -->
    <div v-if="tab==='majors'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Majors / Konsentrasi Keahlian</h3>
      <div class="mt-3 flex flex-wrap gap-2">
        <input v-model="majorForm.code" placeholder="Kode e.g. RPL" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="majorForm.name" placeholder="Nama jurusan" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="majorForm.department_head" placeholder="Kaprog" class="rounded border px-2 py-1.5 text-sm" />
        <button @click="createMajor" class="rounded bg-slate-900 px-3 py-1.5 text-sm text-white">Tambah</button>
      </div>
      <table class="mt-4 w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Kode</th><th class="p-2 text-left">Nama</th><th class="p-2 text-left">Kaprog</th></tr></thead><tbody><tr v-for="m in majors" :key="m.id" class="border-b"><td class="p-2 font-mono">{{ m.code }}</td><td class="p-2">{{ m.name }}</td><td class="p-2">{{ m.department_head }}</td></tr></tbody></table>
    </div>

    <!-- CLASSES -->
    <div v-if="tab==='classes'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Kelas / Rombel</h3>
      <div class="mt-3 flex flex-wrap gap-2">
        <select v-model="classForm.major_id" class="rounded border px-2 py-1.5 text-sm"><option value="">-- jurusan --</option><option v-for="m in majors" :key="m.id" :value="m.id">{{ m.code }} — {{ m.name }}</option></select>
        <select v-model.number="classForm.grade_level" class="rounded border px-2 py-1.5 text-sm"><option :value="10">10</option><option :value="11">11</option><option :value="12">12</option></select>
        <input v-model="classForm.class_name" placeholder="XII RPL 1" class="rounded border px-2 py-1.5 text-sm" />
        <button @click="createClass" class="rounded bg-slate-900 px-3 py-1.5 text-sm text-white">Tambah</button>
      </div>
      <table class="mt-4 w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Kelas</th><th class="p-2 text-left">Jurusan</th><th class="p-2 text-left">Tingkat</th></tr></thead><tbody><tr v-for="c in classes" :key="c.id" class="border-b"><td class="p-2">{{ c.class_name }}</td><td class="p-2">{{ c.major_code }}</td><td class="p-2">{{ c.grade_level }}</td></tr></tbody></table>
    </div>

    <!-- SUBJECTS -->
    <div v-if="tab==='subjects'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Mata Pelajaran</h3>
      <div class="mt-3 flex flex-wrap gap-2">
        <select v-model="subjectForm.major_id" class="rounded border px-2 py-1.5 text-sm"><option value="">-- jurusan (kosong = umum) --</option><option v-for="m in majors" :key="m.id" :value="m.id">{{ m.code }}</option></select>
        <input v-model="subjectForm.code" placeholder="Kode e.g. PWL" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="subjectForm.name" placeholder="Nama mapel" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model.number="subjectForm.pass_grade" type="number" class="w-20 rounded border px-2 py-1.5 text-sm" />
        <select v-model="subjectForm.subject_type" class="rounded border px-2 py-1.5 text-sm"><option value="normatif">normatif</option><option value="adaptif">adaptif</option><option value="produktif">produktif</option></select>
        <button @click="createSubject" class="rounded bg-slate-900 px-3 py-1.5 text-sm text-white">Tambah</button>
      </div>
      <table class="mt-4 w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Kode</th><th class="p-2 text-left">Nama</th><th class="p-2 text-left">Tipe</th><th class="p-2 text-left">KKM</th></tr></thead><tbody><tr v-for="s in subjects" :key="s.id" class="border-b"><td class="p-2 font-mono">{{ s.code }}</td><td class="p-2">{{ s.name }}</td><td class="p-2">{{ s.subject_type }}</td><td class="p-2">{{ s.pass_grade }}</td></tr></tbody></table>
    </div>

    <!-- STUDENTS -->
    <div v-if="tab==='students'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Siswa — password default = NISN</h3>
      <div class="mt-3 flex flex-wrap gap-2">
        <input v-model="studentForm.full_name" placeholder="Nama lengkap" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="studentForm.nisn" placeholder="NISN" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="studentForm.nis" placeholder="NIS (opsional)" class="rounded border px-2 py-1.5 text-sm" />
        <select v-model="studentForm.class_id" class="rounded border px-2 py-1.5 text-sm"><option value="">-- kelas --</option><option v-for="c in classes" :key="c.id" :value="c.id">{{ c.class_name }}</option></select>
        <label class="flex items-center gap-1 text-sm"><input type="checkbox" v-model="studentForm.is_pkl_active" /> PKL</label>
        <button @click="createStudent" class="rounded bg-slate-900 px-3 py-1.5 text-sm text-white">Tambah</button>
      </div>
      <div class="mt-3 overflow-auto"><table class="w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Nama</th><th class="p-2 text-left">NISN</th><th class="p-2 text-left">Kelas</th><th class="p-2 text-left">PKL</th></tr></thead><tbody><tr v-for="s in students" :key="s.id" class="border-b"><td class="p-2">{{ s.full_name }}</td><td class="p-2 font-mono">{{ s.nisn }}</td><td class="p-2">{{ s.class_name }}</td><td class="p-2">{{ s.is_pkl_active ? 'ya' : '' }}</td></tr></tbody></table></div>
    </div>

    <!-- TEACHERS -->
    <div v-if="tab==='teachers'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Guru</h3>
      <div class="mt-3 flex flex-wrap gap-2">
        <input v-model="teacherForm.full_name" placeholder="Nama" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="teacherForm.username" placeholder="username" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="teacherForm.password" type="password" placeholder="password" class="rounded border px-2 py-1.5 text-sm" />
        <input v-model="teacherForm.nip_nuptk" placeholder="NIP/NUPTK" class="rounded border px-2 py-1.5 text-sm" />
        <select v-model="teacherForm.gender" class="rounded border px-2 py-1.5 text-sm"><option value="L">L</option><option value="P">P</option></select>
        <label class="flex items-center gap-1 text-sm"><input type="checkbox" v-model="teacherForm.is_internal" /> internal</label>
        <button @click="createTeacher" class="rounded bg-slate-900 px-3 py-1.5 text-sm text-white">Tambah</button>
      </div>
      <table class="mt-4 w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Nama</th><th class="p-2 text-left">Username</th><th class="p-2 text-left">NIP</th></tr></thead><tbody><tr v-for="t in teachers" :key="t.id" class="border-b"><td class="p-2">{{ t.full_name }}</td><td class="p-2">{{ t.username }}</td><td class="p-2">{{ t.nip_nuptk }}</td></tr></tbody></table>
    </div>

    <!-- QUESTIONS -->
    <div v-if="tab==='questions'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Bank Soal — dukung PG / PG kompleks / matching / short / essay + gambar/audio</h3>
      <div class="mt-3 grid gap-2 sm:grid-cols-2">
        <select v-model="questionForm.subject_id" class="rounded border px-2 py-2 text-sm"><option value="">-- mapel --</option><option v-for="s in subjects" :key="s.id" :value="s.id">{{ s.code }} — {{ s.name }}</option></select>
        <select v-model="questionForm.question_type" class="rounded border px-2 py-2 text-sm"><option value="pg">pg</option><option value="pg_complex">pg_complex</option><option value="matching">matching</option><option value="short">short</option><option value="essay">essay</option></select>
        <textarea v-model="questionForm.content" rows="3" placeholder="Isi soal (bisa diagram text)" class="sm:col-span-2 rounded border px-2 py-2 text-sm"></textarea>
        <input v-model="questionForm.media_url" placeholder="media_url (gambar/audio) opsional" class="rounded border px-2 py-2 text-sm" />
        <select v-model="questionForm.media_type" class="rounded border px-2 py-2 text-sm"><option value="">-- media_type --</option><option value="image">image</option><option value="audio">audio</option></select>
        <textarea v-model="questionForm.optionsText" rows="5" class="sm:col-span-2 rounded border bg-slate-50 px-2 py-2 font-mono text-xs"></textarea>
        <div class="sm:col-span-2 text-xs text-slate-500">PG: [{"id":"A","text":"…","correct":true}]. PG kompleks: multiple correct true. Matching: [{"left":"SIMBOL","right":"FUNGSI"}]. Essay/short: []</div>
        <button @click="createQuestion" class="rounded bg-slate-900 px-3 py-2 text-sm text-white">Simpan Soal</button>
      </div>
      <div class="mt-4 rounded border bg-slate-50 p-3">
        <div class="text-xs font-semibold">Import massal JSON</div>
        <select id="import-subject" class="mt-2 rounded border px-2 py-1.5 text-sm"><option v-for="s in subjects" :key="s.id" :value="s.id">{{ s.code }} — {{ s.name }}</option></select>
        <textarea v-model="importJson" rows="6" placeholder='[{"question_type":"pg","content":"Apa…","options":[{"id":"A","text":"…","correct":true}]}]' class="mt-2 w-full rounded border px-2 py-2 font-mono text-xs"></textarea>
        <button @click="importQuestions" class="mt-2 rounded bg-teal-600 px-3 py-1.5 text-sm text-white">Import</button>
      </div>
      <div class="mt-4 overflow-auto text-sm"><table class="w-full"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Tipe</th><th class="p-2 text-left">Isi</th><th class="p-2 text-left">Opsi</th></tr></thead><tbody><tr v-for="q in filteredQuestions" :key="q.id" class="border-b"><td class="p-2">{{ q.question_type }}</td><td class="p-2 max-w-[28rem] truncate">{{ q.content }}</td><td class="p-2 font-mono text-xs">{{ JSON.stringify(q.options).slice(0,80) }}</td></tr></tbody></table></div>
    </div>

    <!-- EXAMS -->
    <div v-if="tab==='exams'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Sesi Ujian — token auto 6 karakter</h3>
      <div class="mt-3 grid gap-2 sm:grid-cols-2">
        <input v-model="examForm.title" placeholder="Judul e.g. UKK Teori RPL 2026" class="rounded border px-2 py-2 text-sm sm:col-span-2" />
        <select v-model="examForm.subject_id" class="rounded border px-2 py-2 text-sm"><option value="">-- mapel --</option><option v-for="s in subjects" :key="s.id" :value="s.id">{{ s.code }} — {{ s.name }}</option></select>
        <input v-model.number="examForm.duration" type="number" class="rounded border px-2 py-2 text-sm" placeholder="Durasi menit" />
        <label class="flex items-center gap-1 text-sm"><input type="checkbox" v-model="examForm.is_pkl_allowed" /> izinkan PKL remote</label>
        <label class="flex items-center gap-1 text-sm"><input type="checkbox" v-model="examForm.shuffle_questions" /> acak soal</label>
        <label class="flex items-center gap-1 text-sm"><input type="checkbox" v-model="examForm.shuffle_options" /> acak opsi</label>
        <label class="flex items-center gap-1 text-sm"><input type="checkbox" v-model="examForm.is_active" /> aktif</label>
      </div>
      <div class="mt-3">
        <div class="text-xs font-semibold">Target kelas</div>
        <div class="mt-1 flex flex-wrap gap-1"><label v-for="c in classes" :key="c.id" class="rounded border px-2 py-1 text-xs"><input type="checkbox" :value="c.id" v-model="examForm.class_ids" /> {{ c.class_name }}</label></div>
      </div>
      <div class="mt-3">
        <div class="text-xs font-semibold">Pilih soal (order sesuai centang)</div>
        <div class="mt-1 max-h-48 overflow-auto rounded border p-2"><label v-for="q in questions.filter(x=> !examForm.subject_id || x.subject_id===examForm.subject_id)" :key="q.id" class="flex gap-2 border-b py-1 text-xs"><input type="checkbox" :value="q.id" v-model="examForm.question_ids" /> <span class="font-semibold">{{ q.question_type }}</span> {{ q.content.slice(0,80) }}</label></div>
      </div>
      <button @click="createExam" class="mt-3 rounded bg-slate-900 px-4 py-2 text-sm text-white">Buat Ujian</button>

      <div class="mt-6 overflow-auto">
        <table class="w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Judul</th><th class="p-2 text-left">Mapel</th><th class="p-2">Token</th><th class="p-2">Durasi</th><th class="p-2">Soal</th><th></th></tr></thead><tbody><tr v-for="e in exams" :key="e.id" class="border-b"><td class="p-2">{{ e.title }}</td><td class="p-2">{{ e.subject_name }}</td><td class="p-2 font-mono font-bold tracking-widest">{{ e.token }}</td><td class="p-2 text-center">{{ e.duration }}m</td><td class="p-2 text-center">{{ e.question_count }}</td><td class="p-2 flex gap-1"><button @click="toggleExam(e.id)" class="rounded border px-2 py-1 text-xs">{{ e.is_active?'Nonaktif':'Aktifkan' }}</button><button @click="viewResults(e.id); tab='monitor'" class="rounded bg-slate-900 px-2 py-1 text-xs text-white">Hasil</button><button @click="viewMonitor(e.id); tab='monitor'" class="rounded border px-2 py-1 text-xs">Monitor</button></td></tr></tbody></table>
      </div>
    </div>

    <!-- MONITOR / RESULTS -->
    <div v-if="tab==='monitor'" class="rounded-xl border bg-white p-4">
      <h3 class="font-semibold">Monitoring & Rekap — pilih ujian di tab Exams dulu, atau pilih di bawah:</h3>
      <select v-model="selectedExam" @change="viewResults(selectedExam)" class="mt-2 rounded border px-2 py-1.5 text-sm"><option value="">-- pilih ujian --</option><option v-for="e in exams" :key="e.id" :value="e.id">{{ e.title }} — {{ e.token }}</option></select>
      <div class="mt-3 flex gap-2">
        <button @click="viewResults(selectedExam)" :disabled="!selectedExam" class="rounded bg-slate-900 px-3 py-1.5 text-sm text-white disabled:opacity-40">Muat Hasil</button>
        <button @click="viewMonitor(selectedExam)" :disabled="!selectedExam" class="rounded border px-3 py-1.5 text-sm">Muat Monitor (live)</button>
        <button @click="gradeExam(selectedExam)" :disabled="!selectedExam" class="rounded border px-3 py-1.5 text-sm">Grade Ulang (auto PG)</button>
        <button @click="exportCsv" :disabled="!results.length" class="rounded bg-teal-600 px-3 py-1.5 text-sm text-white disabled:opacity-40">Export CSV</button>
      </div>

      <div v-if="monitor.length" class="mt-4 overflow-auto">
        <div class="mb-2 text-xs font-semibold">Live monitor</div>
        <table class="w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Siswa</th><th class="p-2">NISN</th><th class="p-2">Kelas</th><th class="p-2">Status</th><th class="p-2">Terjawab</th><th class="p-2">Online</th><th class="p-2">Cheat</th><th class="p-2">Nilai</th></tr></thead><tbody><tr v-for="m in monitor" :key="m.student_id" class="border-b"><td class="p-2">{{ m.student_name }}</td><td class="p-2 font-mono">{{ m.nisn }}</td><td class="p-2">{{ m.class_name }}</td><td class="p-2"><span :class="m.status==='submitted'?'bg-teal-600 text-white': m.status==='graded'?'bg-slate-900 text-white':'bg-slate-100'" class="rounded px-1.5 py-0.5 text-xs">{{ m.status }}</span></td><td class="p-2 text-center">{{ m.answered }}/{{ m.total }}</td><td class="p-2 text-center">{{ m.online ? '●' : '○' }}</td><td class="p-2 text-center">{{ m.cheat_warnings }}</td><td class="p-2 text-center font-bold">{{ m.final_score }}</td></tr></tbody></table>
      </div>

      <div v-if="results.length" class="mt-4 overflow-auto">
        <div class="mb-2 text-xs font-semibold">Rekap nilai per siswa</div>
        <table class="w-full text-sm"><thead><tr class="border-b text-slate-500"><th class="p-2 text-left">Siswa</th><th class="p-2">NISN</th><th class="p-2">Kelas</th><th class="p-2">PG</th><th class="p-2">Essay</th><th class="p-2">Akhir</th><th class="p-2">Terjawab</th><th class="p-2">Status</th></tr></thead><tbody><tr v-for="r in results" :key="r.id" class="border-b"><td class="p-2">{{ r.student_name }}</td><td class="p-2 font-mono">{{ r.nisn }}</td><td class="p-2">{{ r.class_name }}</td><td class="p-2 text-center">{{ r.total_pg_score }}</td><td class="p-2 text-center">{{ r.total_essay_score }}</td><td class="p-2 text-center font-bold">{{ r.final_score }}</td><td class="p-2 text-center">{{ r.answered_count }}</td><td class="p-2">{{ r.status }}</td></tr></tbody></table>
      </div>
      <p v-if="selectedExam && !results.length && !monitor.length" class="mt-3 text-sm text-slate-500">Belum ada data hasil untuk ujian ini — atau siswa belum mulai.</p>
    </div>
  </div>
</template>
