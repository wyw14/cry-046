import { defineStore } from 'pinia'
import { listProjects, toggleFavorite, type Project } from '../api/services'
export const useProjectStore = defineStore('projects', { state: () => ({ items: [] as Project[], q: '', loading: false }), actions: { async refresh() { this.loading = true; try { this.items = (await listProjects(this.q)).items } finally { this.loading = false } }, async favorite(id: string) { await toggleFavorite(id) } } })
