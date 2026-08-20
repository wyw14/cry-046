import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SeverityTag from '@/components/SeverityTag.vue'
import PageHeader from '@/components/PageHeader.vue'
import PaginationBar from '@/components/PaginationBar.vue'
import ToastStack from '@/components/ToastStack.vue'
import { setActivePinia, createPinia } from 'pinia'
import { useNotificationsStore } from '@/stores/notifications'

describe('SeverityTag', () => {
  it('renders severity', () => {
    const w = mount(SeverityTag, { props: { severity: 'critical' } })
    expect(w.text()).toBe('critical')
    expect(w.classes()).toContain('tag-critical')
  })

  it('renders status', () => {
    const w = mount(SeverityTag, { props: { status: 'resolved' } })
    expect(w.text()).toBe('resolved')
    expect(w.classes()).toContain('tag-resolved')
  })

  it('renders nothing when no prop', () => {
    const w = mount(SeverityTag)
    expect(w.text()).toBe('')
  })
})

describe('PageHeader', () => {
  it('renders title and description', () => {
    const w = mount(PageHeader, { props: { title: '工作台', description: '欢迎' } })
    expect(w.text()).toContain('工作台')
    expect(w.text()).toContain('欢迎')
  })

  it('renders actions slot', () => {
    const w = mount(PageHeader, {
      props: { title: '页面' },
      slots: { actions: '<button class="btn">操作</button>' },
    })
    expect(w.html()).toContain('操作')
  })
})

describe('PaginationBar', () => {
  it('emits change with previous page', async () => {
    const w = mount(PaginationBar, {
      props: { page: { page: 2, page_size: 20, total: 40, has_next: true } },
    })
    const buttons = w.findAll('button')
    await buttons[0].trigger('click')
    expect(w.emitted('change')?.[0]).toEqual([1])
  })

  it('emits change with next page', async () => {
    const w = mount(PaginationBar, {
      props: { page: { page: 1, page_size: 20, total: 40, has_next: true } },
    })
    const buttons = w.findAll('button')
    await buttons[1].trigger('click')
    expect(w.emitted('change')?.[0]).toEqual([2])
  })

  it('disables prev on first page', () => {
    const w = mount(PaginationBar, {
      props: { page: { page: 1, page_size: 20, total: 40, has_next: true } },
    })
    const buttons = w.findAll('button')
    expect(buttons[0].attributes('disabled')).toBeDefined()
  })

  it('disables next when no next page', () => {
    const w = mount(PaginationBar, {
      props: { page: { page: 2, page_size: 20, total: 40, has_next: false } },
    })
    const buttons = w.findAll('button')
    expect(buttons[1].attributes('disabled')).toBeDefined()
  })
})

describe('ToastStack', () => {
  it('renders notifications from store', async () => {
    setActivePinia(createPinia())
    const store = useNotificationsStore()
    store.push('info', '一条')
    store.push('error', '另一条')
    const w = mount(ToastStack)
    expect(w.findAll('.toast')).toHaveLength(2)
    expect(w.text()).toContain('一条')
    expect(w.text()).toContain('另一条')
  })

  it('dismisses on click', async () => {
    setActivePinia(createPinia())
    const store = useNotificationsStore()
    store.push('info', 'click me')
    const w = mount(ToastStack)
    await w.find('.toast').trigger('click')
    expect(w.findAll('.toast')).toHaveLength(0)
  })
})
