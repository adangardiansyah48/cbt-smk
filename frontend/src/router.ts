import { createRouter, createWebHistory } from "vue-router"
import { token } from "./lib/api"

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: "/login" },
    { path: "/login", component: () => import("./pages/Login.vue"), meta: { public: true } },
    { path: "/exam", component: () => import("./pages/Exam.vue"), meta: { role: "student" } },
    { path: "/admin", component: () => import("./pages/Admin.vue"), meta: { role: "staff" } },
  ],
})

router.beforeEach((to) => {
  const pub = (to.meta as any).public
  if (pub) return true
  if (!token()) return { path: "/login" }
  const role = localStorage.getItem("cbt_role") || ""
  const need = (to.meta as any).role as string | undefined
  if (need === "student" && role !== "student") return { path: "/login" }
  if (need === "staff" && !["admin", "teacher"].includes(role)) return { path: "/login" }
  return true
})
