import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useNotificationsStore } from '@/stores/notifications'

describe('notifications store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('push adds a notification', () => {
    const store = useNotificationsStore()
    store.push('info', 'hello')
    expect(store.items).toHaveLength(1)
    expect(store.items[0].message).toBe('hello')
    expect(store.items[0].kind).toBe('info')
  })

  it('dismiss removes by id', () => {
    const store = useNotificationsStore()
    const id = store.push('error', 'oops')
    store.push('info', 'kept')
    store.dismiss(id)
    expect(store.items).toHaveLength(1)
    expect(store.items[0].message).toBe('kept')
  })

  it('auto-dismisses after ttl', () => {
    const store = useNotificationsStore()
    store.push('info', 'temp')
    expect(store.items).toHaveLength(1)
    vi.advanceTimersByTime(5000)
    expect(store.items).toHaveLength(0)
  })

  it('clear removes all', () => {
    const store = useNotificationsStore()
    store.push('info', 'a')
    store.push('error', 'b')
    store.clear()
    expect(store.items).toHaveLength(0)
  })
})
