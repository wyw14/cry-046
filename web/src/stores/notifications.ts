import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Notification {
  id: number
  kind: 'success' | 'error' | 'info'
  message: string
  ttl: number
}

let counter = 0

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<Notification[]>([])

  function push(kind: Notification['kind'], message: string) {
    counter += 1
    const id = counter
    items.value.push({ id, kind, message, ttl: 4000 })
    setTimeout(() => dismiss(id), 4000)
    return id
  }

  function dismiss(id: number) {
    const idx = items.value.findIndex((n) => n.id === id)
    if (idx >= 0) items.value.splice(idx, 1)
  }

  function clear() {
    items.value = []
  }

  return { items, push, dismiss, clear }
})
