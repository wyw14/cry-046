import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { setTenant, setToken, getTenant, getToken } from '@/api/http'

// 用户与租户的会话状态. 离线演示使用 token 即用户名,
// 后端通过 token 映射到预设的演示账号 (admin/operator/assignee/reviewer).
export type Role = 'admin' | 'operator' | 'assignee' | 'reviewer'

export interface SessionUser {
  username: string
  role: Role
  tenant: string
}

const ROLE_BY_TOKEN: Record<string, Role> = {
  admin: 'admin',
  operator: 'operator',
  assignee: 'assignee',
  reviewer: 'reviewer',
}

export const useSessionStore = defineStore('session', () => {
  const token = ref<string>(getToken())
  const tenant = ref<string>(getTenant())

  const user = computed<SessionUser | null>(() => {
    if (!token.value) return null
    const role = ROLE_BY_TOKEN[token.value]
    if (!role) return null
    return { username: token.value, role, tenant: tenant.value }
  })

  const isAuthenticated = computed(() => user.value !== null)

  function login(t: string, ten: string = 'default') {
    token.value = t
    tenant.value = ten
    setToken(t)
    setTenant(ten)
  }

  function logout() {
    token.value = ''
    setToken('')
  }

  function setTenantID(t: string) {
    tenant.value = t
    setTenant(t)
  }

  return { token, tenant, user, isAuthenticated, login, logout, setTenantID }
})
