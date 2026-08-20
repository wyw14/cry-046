import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
import LoginView from '@/views/LoginView.vue'

function mountLogin(redirect = '/') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
      { path: '/', name: 'dashboard', component: { template: '<div>dashboard</div>' } },
    ],
  })
  router.push(`/login?redirect=${redirect}`)
  const wrapper = mount(LoginView, { global: { plugins: [router, createPinia()] } })
  return { wrapper, router }
}

describe('LoginView', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renders inputs and quick buttons', () => {
    const { wrapper } = mountLogin()
    expect(wrapper.find('input.input').exists()).toBe(true)
    const buttons = wrapper.findAll('button')
    const quickButtons = buttons.filter((b) => ['admin', 'operator', 'assignee', 'reviewer'].includes(b.text()))
    expect(quickButtons).toHaveLength(4)
  })

  it('shows error when token empty', async () => {
    const { wrapper } = mountLogin()
    await wrapper.find('input.input').setValue('')
    await wrapper.find('button.btn-primary').trigger('click')
    expect(wrapper.text()).toContain('请输入演示令牌')
  })

  it('quick login redirects to dashboard', async () => {
    const { wrapper, router } = mountLogin()
    const buttons = wrapper.findAll('button')
    const adminBtn = buttons.find((b) => b.text() === 'admin')
    await adminBtn?.trigger('click')
    await flushPromises()
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('dashboard')
    expect(localStorage.getItem('wsr.token')).toBe('admin')
  })

  it('login with custom token stores token', async () => {
    const { wrapper, router } = mountLogin()
    await wrapper.find('input.input').setValue('operator')
    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('dashboard')
    expect(localStorage.getItem('wsr.token')).toBe('operator')
  })
})
