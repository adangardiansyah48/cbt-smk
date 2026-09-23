import { reactive } from "vue"
import { apiFetch, setToken, clearToken } from "./api"

export const auth = reactive<{ role: string; claims: any; ready: boolean }>({
  role: localStorage.getItem("cbt_role") || "",
  claims: null,
  ready: false,
})

export async function loginNISN(nisn: string, examToken: string) {
  const res = await apiFetch("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ nisn, token: examToken }),
  })
  setToken(res.token)
  localStorage.setItem("cbt_role", res.role)
  auth.role = res.role
  auth.claims = res.student
  return res
}

export async function loginStaff(username: string, password: string) {
  const res = await apiFetch("/api/auth/staff", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  })
  setToken(res.token)
  localStorage.setItem("cbt_role", res.role)
  auth.role = res.role
  auth.claims = res.user
  return res
}

export function logout() {
  clearToken()
  localStorage.removeItem("cbt_role")
  auth.role = ""
  auth.claims = null
}

export async function fetchMe() {
  try {
    const m = await apiFetch("/api/me")
    auth.claims = m
    auth.role = m.role || auth.role
  } catch (e) {
    // ignore
  } finally {
    auth.ready = true
  }
}
