<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue"
import { apiFetch } from "../lib/api"
import { logout } from "../lib/auth"
import { useRouter } from "vue-router"

const router = useRouter()
type Q = { id:string; question_type:string; content:string; media_url?:string; media_type?:string; options:any; order_no:number }
type Paper = { exam:any; questions:Q[]; saved: Record<string,{answer:any;is_flagged:boolean}>; remaining_sec:number; started_at:string }

const paper = ref<Paper | null>(null)
const idx = ref(0)
const answers = ref<Record<string, any>>({})
const flagged = ref<Record<string, boolean>>({})
const remain = ref(0)
const err = ref("")
const saving = ref<Record<string,boolean>>({})
const submitted = ref(false)
const score = ref<any>(null)
let timer: any = null
let hbTimer: any = null
let cheatCount = 0
let localKey = ""

const current = computed(() => paper.value?.questions[idx.value])
const total = computed(() => paper.value?.questions.length || 0)
const answeredCount = computed(() => Object.keys(answers.value).filter(k => answers.value[k]!=null && answers.value[k]!=='' && String(answers.value[k])!=='').length)
const mm = computed(() => String(Math.floor(remain.value/60)).padStart(2,"0"))
const ss = computed(() => String(remain.value%60).padStart(2,"0"))

function fmt(sec:number){ return `${String(Math.floor(sec/60)).padStart(2,"0")}:${String(sec%60).padStart(2,"0")}` }

async function load() {
  try {
    const p: Paper = await apiFetch("/api/exam/paper")
    paper.value = p
    remain.value = p.remaining_sec ?? p.exam.duration * 60
    // restore saved + localStorage
    localKey = `cbt:${p.exam.id}:answers`
    for (const [qid, v] of Object.entries(p.saved || {})) { answers.value[qid] = v.answer; flagged.value[qid]=!!v.is_flagged }
    try {
      const ls = JSON.parse(localStorage.getItem(localKey) || "{}")
      for (const [k,v] of Object.entries(ls)) if (answers.value[k]==null) answers.value[k]=v
    } catch {}
    startTimers()
    installAntiCheat()
  } catch (e: any) { err.value = e.message }
}

function startTimers() {
  timer = setInterval(() => {
    if (remain.value > 0) remain.value--
    else { submit(true) }
  }, 1000)
  hbTimer = setInterval(async () => {
    try { await apiFetch("/api/exam/heartbeat", { method: "POST" }) } catch {}
  }, 5000)
}

function saveLocal() {
  try { localStorage.setItem(localKey, JSON.stringify(answers.value)); localStorage.setItem(localKey+":flags", JSON.stringify(flagged.value)) } catch {}
}

async function upsert(qid: string) {
  saveLocal()
  saving.value[qid]=true
  const payload = { question_id: qid, answer: answers.value[qid] ?? null, is_flagged: !!flagged.value[qid] }
  try { await apiFetch("/api/exam/response", { method:"POST", body: JSON.stringify(payload) }) }
  catch (e) { /* keep in local */ }
  finally { saving.value[qid]=false }
}

function onPg(qid:string, optId:string) { answers.value[qid]=optId; upsert(qid) }
function onPgComplex(qid:string, optId:string) {
  const cur: string[] = Array.isArray(answers.value[qid]) ? answers.value[qid] : []
  const next = cur.includes(optId) ? cur.filter(x=>x!==optId) : [...cur, optId]
  answers.value[qid]=next; upsert(qid)
}
function onMatching(qid:string, left:string, right:string) {
  const cur: Record<string,string> = typeof answers.value[qid]==='object' && answers.value[qid] && !Array.isArray(answers.value[qid]) ? answers.value[qid] : {}
  answers.value[qid]={ ...cur, [left]: right }; upsert(qid)
}
let debounce:any=null
function onEssayInput(qid:string) {
  clearTimeout(debounce); debounce=setTimeout(()=> upsert(qid), 800)
}
function toggleFlag(qid:string){ flagged.value[qid]=!flagged.value[qid]; upsert(qid) }

async function submit(auto=false) {
  if (submitted.value) return
  if (!auto && !confirm("Kumpulkan jawaban? Tidak bisa diubah lagi.")) return
  try {
    const res = await apiFetch("/api/exam/submit", { method:"POST" })
    score.value=res; submitted.value=true
    cleanup()
  } catch (e:any) { err.value=e.message }
}

function installAntiCheat() {
  const onVisibility = async () => {
    if (document.hidden) {
      cheatCount++
      try { await apiFetch("/api/exam/heartbeat?cheat=1", { method:"POST" }) } catch {}
      const msg = `Peringatan ${cheatCount}: jangan pindah tab/aplikasi.`
      alert(msg)
    }
  }
  document.addEventListener("visibilitychange", onVisibility)
  document.addEventListener("contextmenu", e => e.preventDefault())
  document.addEventListener("copy", e => e.preventDefault())
  document.addEventListener("cut", e => e.preventDefault())
  document.addEventListener("beforeunload", () => { saveLocal() })
}

function cleanup() { clearInterval(timer); clearInterval(hbTimer); if (localKey) localStorage.removeItem(localKey) }
onBeforeUnmount(cleanup)
onMounted(load)

function doLogout(){ cleanup(); logout(); router.push("/login") }
</script>

<template>
  <div class="min-h-dvh">
    <div v-if="submitted" class="mx-auto max-w-2xl p-6">
      <div class="rounded-2xl border bg-white p-8 text-center shadow-sm">
        <h2 class="text-xl font-bold text-teal-700">Ujian selesai — jawaban terkumpul</h2>
        <p class="mt-2 text-sm text-slate-600">Skor PG: <b>{{ score?.total_pg_score?.toFixed(1) }}</b> · Nilai akhir: <b>{{ score?.final_score?.toFixed(1) }}</b></p>
        <p class="mt-1 text-xs text-slate-400">Rincian nilai lengkap ada di dashboard guru. Essay menunggu penilaian manual.</p>
        <button @click="doLogout" class="mt-6 rounded-lg bg-slate-900 px-6 py-2.5 text-sm text-white">Keluar</button>
      </div>
    </div>

    <div v-else-if="!paper" class="mx-auto max-w-3xl p-8">
      <p v-if="err" class="rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ err }}</p>
      <p v-else class="text-sm text-slate-500">Memuat soal…</p>
    </div>

    <template v-else>
      <div class="mx-auto flex max-w-6xl gap-4 p-4">
        <div class="flex-1">
          <div class="mb-3 flex items-center justify-between rounded-xl border bg-white px-4 py-3">
            <div>
              <div class="text-sm font-semibold">{{ paper.exam.title }} — {{ paper.exam.subject_name }}</div>
              <div class="text-xs text-slate-500">Soal {{ idx+1 }}/{{ total }} · Terjawab {{ answeredCount }}/{{ total }}</div>
            </div>
            <div class="rounded-full bg-slate-900 px-3 py-1.5 font-mono text-sm font-bold text-white">{{ mm }}:{{ ss }}</div>
          </div>

          <div v-if="current" class="rounded-xl border bg-white p-5 shadow-sm">
            <div class="mb-2 flex items-center gap-2">
              <span class="rounded bg-teal-600 px-2 py-0.5 text-xs font-bold text-white">{{ current.question_type.toUpperCase() }}</span>
              <span class="text-xs text-slate-400">#{{ idx+1 }}</span>
              <button @click="toggleFlag(current.id)" :class="flagged[current.id] ? 'bg-amber-500 text-white' : 'border text-slate-600'" class="ml-auto rounded px-2 py-1 text-xs">
                {{ flagged[current.id] ? '★ Ragu' : 'Tandai ragu' }}
              </button>
              <span v-if="saving[current.id]" class="text-xs text-slate-400">menyimpan…</span>
            </div>
            <div class="prose prose-sm max-w-none">{{ current.content }}</div>
            <img v-if="current.media_type==='image' && current.media_url" :src="current.media_url" alt="media" class="mt-3 max-h-72 rounded border" />
            <audio v-if="current.media_type==='audio' && current.media_url" :src="current.media_url" controls class="mt-3 w-full" />

            <!-- PG -->
            <div v-if="current.question_type==='pg'" class="mt-4 space-y-2">
              <label v-for="opt in current.options" :key="opt.id" class="flex cursor-pointer items-center gap-3 rounded-lg border p-3 hover:bg-slate-50" :class="answers[current.id]===opt.id ? 'border-teal-600 bg-teal-50' : ''">
                <input type="radio" :name="current.id" :value="opt.id" :checked="answers[current.id]===opt.id" @change="onPg(current!.id, opt.id)" />
                <span class="text-sm">{{ opt.text || opt.id }}</span>
              </label>
            </div>

            <!-- PG complex -->
            <div v-else-if="current.question_type==='pg_complex'" class="mt-4 space-y-2">
              <p class="text-xs text-slate-500">Pilih semua yang benar:</p>
              <label v-for="opt in current.options" :key="opt.id" class="flex cursor-pointer items-center gap-3 rounded-lg border p-3 hover:bg-slate-50">
                <input type="checkbox" :checked="(answers[current.id]||[]).includes(opt.id)" @change="onPgComplex(current!.id, opt.id)" />
                <span class="text-sm">{{ opt.text || opt.id }}</span>
              </label>
            </div>

            <!-- Matching -->
            <div v-else-if="current.question_type==='matching'" class="mt-4 space-y-2">
              <p class="text-xs text-slate-500">Jodohkan left → right:</p>
              <div v-for="pair in current.options" :key="pair.left" class="flex items-center gap-2">
                <span class="w-40 rounded bg-slate-100 px-3 py-2 text-sm">{{ pair.left }}</span>
                <span class="text-slate-400">→</span>
                <select :value="(answers[current.id]||{})[pair.left] || ''" @change="onMatching(current!.id, pair.left, ($event.target as HTMLSelectElement).value)" class="flex-1 rounded-lg border px-2 py-2 text-sm">
                  <option value="">-- pilih --</option>
                  <option v-for="o in current.options" :key="o.right" :value="o.right">{{ o.right }}</option>
                </select>
              </div>
            </div>

            <!-- Short / Essay -->
            <div v-else-if="current.question_type==='short' || current.question_type==='essay'" class="mt-4">
              <textarea v-model="answers[current.id]" @input="onEssayInput(current.id)" rows="5" :placeholder="current.question_type==='short' ? 'Jawaban singkat…' : 'Uraian jawaban…'" class="w-full rounded-lg border p-3 text-sm"></textarea>
              <p class="mt-1 text-xs text-slate-400">Auto-save tiap mengetik (debounce 0.8s + LocalStorage tiap 1 detik).</p>
            </div>

            <div class="mt-5 flex justify-between">
              <button @click="idx=Math.max(0, idx-1)" :disabled="idx===0" class="rounded-lg border px-4 py-2 text-sm disabled:opacity-40">← Sebelumnya</button>
              <button v-if="idx < total-1" @click="idx++" class="rounded-lg bg-slate-900 px-6 py-2 text-sm text-white">Selanjutnya →</button>
              <button v-else @click="submit(false)" class="rounded-lg bg-teal-600 px-8 py-2 text-sm font-semibold text-white">Kumpulkan</button>
            </div>
          </div>

          <p v-if="err" class="mt-3 text-sm text-red-600">{{ err }}</p>
        </div>

        <aside class="hidden w-64 shrink-0 lg:block">
          <div class="rounded-xl border bg-white p-3">
            <div class="mb-2 text-xs font-semibold text-slate-700">Navigasi — {{ fmt(remain) }} sisa</div>
            <div class="grid grid-cols-5 gap-1.5">
              <button v-for="(q,i) in paper.questions" :key="q.id" @click="idx=i" class="rounded-md px-1.5 py-2 text-xs font-medium"
                :class="i===idx ? 'bg-slate-900 text-white' : answers[q.id]!=null && answers[q.id]!=='' ? 'bg-teal-600 text-white' : flagged[q.id] ? 'bg-amber-400 text-slate-900' : 'bg-slate-100 text-slate-600'">
                {{ i+1 }}
              </button>
            </div>
            <div class="mt-3 flex flex-wrap gap-1 text-[10px]">
              <span class="rounded bg-teal-600 px-1.5 py-0.5 text-white">Terjawab</span>
              <span class="rounded bg-slate-100 px-1.5 py-0.5">Kosong</span>
              <span class="rounded bg-amber-400 px-1.5 py-0.5">Ragu</span>
            </div>
            <button @click="submit(false)" class="mt-4 w-full rounded-lg bg-teal-600 py-2.5 text-sm font-semibold text-white">Kumpulkan Sekarang</button>
          </div>
        </aside>
      </div>
    </template>
  </div>
</template>
