import { api } from './client'
export type Project = { id: string; name: string; customer: string; series: string; state: string }
export const listProjects = (q = '') => api<{items: Project[]; total: number}>(`/projects?q=${encodeURIComponent(q)}`)
export const archiveProject = (id: string) => api<Project>(`/projects/${id}/archive`, { method: 'POST' })
export const toggleFavorite = (id: string) => api<{favorite: boolean}>(`/projects/${id}/favorite`, { method: 'POST' })
