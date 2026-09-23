<script setup lang="ts">
import { onMounted } from "vue"
import { useRoute, useRouter } from "vue-router"
import { auth, logout } from "./lib/auth"
import { fetchMe } from "./lib/auth"

const route = useRoute()
const router = useRouter()

onMounted(() => fetchMe())

function doLogout() {
  logout()
  router.push("/login")
}
</script>

<template>
  <div class="min-h-dvh bg-slate-50 text-slate-900">
    <header v-if="!['/login','/'].includes(route.path)" class="sticky top-0 z-20 border-b bg-white/80 backdrop-blur">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
        <div class="flex items-center gap-2">
          <span class="rounded bg-teal-600 px-2 py-1 text-sm font-bold text-white">CBT SMK</span>
          <span class="text-sm text-slate-500">
            {{ auth.role === 'student' ? 'Mode Siswa' : 'Admin / Guru' }}
          </span>
        </div>
        <button @click="doLogout" class="rounded bg-slate-900 px-3 py-1.5 text-sm text-white hover:bg-slate-800">Keluar</button>
      </div>
    </header>
    <main>
      <router-view />
    </main>
  </div>
</template>
