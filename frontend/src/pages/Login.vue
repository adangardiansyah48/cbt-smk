<script setup lang="ts">
import { ref, onMounted } from "vue"
import { useRouter } from "vue-router"
import { apiFetch } from "../lib/api"
import { loginNISN, loginStaff } from "../lib/auth"

const router = useRouter()
const tab = ref<"student" | "staff">("student")
const nisn = ref("")
const examToken = ref("")
const username = ref("")
const password = ref("")
const err = ref("")
const busy = ref(false)
const school = ref<any>(null)

onMounted(async () => {
  try { school.value = await apiFetch("/api/school") } catch {}
})

async function doStudent() {
  err.value = ""; busy.value = true
  try {
    await loginNISN(nisn.value.trim(), examToken.value.trim())
    router.push("/exam")
  } catch (e: any) { err.value = e.message } finally { busy.value = false }
}
async function doStaff() {
  err.value = ""; busy.value = true
  try {
    await loginStaff(username.value.trim(), password.value)
    router.push("/admin")
  } catch (e: any) { err.value = e.message } finally { busy.value = false }
}
</script>

<template>
  <div class="mx-auto flex min-h-dvh max-w-md flex-col justify-center gap-6 p-6">
    <div class="text-center">
      <div class="mx-auto mb-2 w-fit rounded-lg bg-teal-600 px-3 py-1.5 text-sm font-bold text-white">CBT SMK ANBK & TKA</div>
      <h1 class="text-xl font-bold text-slate-900">{{ school?.school_name || 'Sistem Ujian CBT Terpadu' }}</h1>
      <p class="mt-1 text-sm text-slate-500">Login dengan NISN + Token Ujian (siswa) atau akun Guru/Admin</p>
    </div>

    <div class="flex rounded-lg border bg-white p-1">
      <button @click="tab='student'" :class="tab==='student' ? 'bg-slate-900 text-white' : 'text-slate-600'" class="flex-1 rounded-md py-2 text-sm font-medium">Siswa (NISN)</button>
      <button @click="tab='staff'" :class="tab==='staff' ? 'bg-slate-900 text-white' : 'text-slate-600'" class="flex-1 rounded-md py-2 text-sm font-medium">Guru / Admin</button>
    </div>

    <div v-if="tab==='student'" class="rounded-2xl border bg-white p-5 shadow-sm">
      <div class="space-y-3">
        <label class="block text-sm font-medium">NISN</label>
        <input v-model="nisn" placeholder="e.g. 0051234567" class="w-full rounded-lg border px-3 py-2.5 text-sm" />
        <label class="block text-sm font-medium">Token Ujian</label>
        <input v-model="examToken" placeholder="6 karakter, e.g. A1B2C3" class="w-full rounded-lg border px-3 py-2.5 text-sm uppercase tracking-widest" />
        <p v-if="err" class="text-sm text-red-600">{{ err }}</p>
        <button @click="doStudent" :disabled="busy" class="w-full rounded-lg bg-teal-600 py-3 text-sm font-semibold text-white disabled:opacity-50">Masuk Ujian</button>
        <p class="text-center text-xs text-slate-400">Mode PKL: jika status PKL aktif, ujian remote auto-diizinkan bila token mengizinkan.</p>
      </div>
    </div>

    <div v-else class="rounded-2xl border bg-white p-5 shadow-sm">
      <div class="space-y-3">
        <label class="block text-sm font-medium">Username</label>
        <input v-model="username" placeholder="admin / username guru" class="w-full rounded-lg border px-3 py-2.5 text-sm" />
        <label class="block text-sm font-medium">Password</label>
        <input v-model="password" type="password" class="w-full rounded-lg border px-3 py-2.5 text-sm" />
        <p v-if="err" class="text-sm text-red-600">{{ err }}</p>
        <button @click="doStaff" :disabled="busy" class="w-full rounded-lg bg-slate-900 py-3 text-sm font-semibold text-white disabled:opacity-50">Masuk Admin</button>
        <p class="text-center text-xs text-slate-400">Default: admin / admin123 — ganti setelah install.</p>
      </div>
    </div>
  </div>
</template>
