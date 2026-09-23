export const API = import.meta.env.VITE_API_URL || "http://localhost:8080"

export function token() {
  return localStorage.getItem("cbt_token") || ""
}
export function setToken(t: string) {
  localStorage.setItem("cbt_token", t)
}
export function clearToken() {
  localStorage.removeItem("cbt_token")
}
export function authHeaders(extra: Record<string, string> = {}): Record<string, string> {
  const t = token()
  return t ? { Authorization: `Bearer ${t}`, ...extra } : extra
}
export async function apiFetch(path: string, init: RequestInit = {}) {
  const res = await fetch(`${API}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...authHeaders(), ...(init.headers as any) },
  })
  const body = await res.json().catch(() => ({}))
  if (!res.ok) throw Object.assign(new Error(body.error || res.statusText), { status: res.status, body })
  return body
}
