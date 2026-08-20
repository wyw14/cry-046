import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
import { useSessionStore } from '@/stores/session'
import AppLayout from '@/views/AppLayout.vue'

describe('AppLayout navigation', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
  })

  it('renders sidebar links', () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: AppLayout, children: [
          { path: '', name: 'dashboard', component: { template: '<div>dashboard</div>' } },
        ] },
      ],
    })
    const session = useSessionStore()
    session.login('admin', 'default')
    const wrapper = mount(AppLayout, { global: { plugins: [router, createPinia()] } })
    expect(wrapper.html()).toContain('工作台')
    expect(wrapper.html()).toContain('项目')
    expect(wrapper.html()).toContain('异常处置')
    expect(wrapper.html()).toContain('汇总与重算')
    expect(wrapper.html()).toContain('审计与导出')
  })

  it('shows tenant and user info', () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', component: AppLayout, children: [
          { path: '', name: 'dashboard', component: { template: '<div></div>' } },
        ] },
      ],
    })
    const session = useSessionStore()
    session.login('reviewer', 'demo-tenant')
    const wrapper = mount(AppLayout, { global: { plugins: [router, createPinia()] } })
    expect(wrapper.html()).toContain('demo-tenant')
    expect(wrapper.html()).toContain('reviewer')
  })
})
