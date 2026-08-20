import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useSessionStore } from '@/stores/session'

describe('session store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('starts unauthenticated', () => {
    const store = useSessionStore()
    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
  })

  it('login sets token and tenant', () => {
    const store = useSessionStore()
    store.login('admin', 'demo-tenant')
    expect(store.token).toBe('admin')
    expect(store.tenant).toBe('demo-tenant')
    expect(store.user?.username).toBe('admin')
    expect(store.user?.role).toBe('admin')
    expect(localStorage.getItem('wsr.token')).toBe('admin')
    expect(localStorage.getItem('wsr.tenant')).toBe('demo-tenant')
  })

  it('logout clears token', () => {
    const store = useSessionStore()
    store.login('operator', 'default')
    store.logout()
    expect(store.token).toBe('')
    expect(store.isAuthenticated).toBe(false)
    expect(localStorage.getItem('wsr.token')).toBe('')
  })

  it('unknown token yields null user', () => {
    const store = useSessionStore()
    store.login('bogus', 'default')
    expect(store.user).toBeNull()
    expect(store.isAuthenticated).toBe(false)
  })

  it('switching tenant updates state and localStorage', () => {
    const store = useSessionStore()
    store.login('admin', 'default')
    store.setTenantID('other')
    expect(store.tenant).toBe('other')
    expect(localStorage.getItem('wsr.tenant')).toBe('other')
  })
})
