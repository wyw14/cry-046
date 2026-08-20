// Test setup: provide a working localStorage shim in jsdom.
// In some jsdom/vitest combinations the global localStorage is exposed as a
// plain object missing the Storage prototype methods (clear in particular),
// which breaks tests that call localStorage.clear(). We replace it with a
// minimal in-memory implementation that has the full Storage interface.
const g = globalThis as any
const store = new Map<string, string>()
const shim = {
  get length() {
    return store.size
  },
  clear() {
    store.clear()
  },
  getItem(key: string) {
    return store.has(key) ? store.get(key)! : null
  },
  setItem(key: string, value: string) {
    store.set(key, String(value))
  },
  removeItem(key: string) {
    store.delete(key)
  },
  key(index: number) {
    const keys = Array.from(store.keys())
    return index >= 0 && index < keys.length ? keys[index] : null
  },
}
g.localStorage = shim
g.sessionStorage = shim
