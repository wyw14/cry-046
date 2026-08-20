<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { AuditService } from '@/api/services'
import type { AuditEntry, PageResult } from '@/types'
import PageHeader from '@/components/PageHeader.vue'
import PaginationBar from '@/components/PaginationBar.vue'
import { useNotificationsStore } from '@/stores/notifications'

const notifications = useNotificationsStore()
const items = ref<AuditEntry[]>([])
const page = ref<PageResult>({ page: 1, page_size: 50, total: 0, has_next: false })
const loading = ref(false)
const filters = reactive({ actor_id: '', action: '', entity_id: '' })

async function load() {
  loading.value = true
  try {
    const res = await AuditService.list({
      page: page.value.page,
      page_size: page.value.page_size,
      actor_id: filters.actor_id || undefined,
      action: filters.action || undefined,
      entity_id: filters.entity_id || undefined,
    })
    items.value = res.items
    page.value = { page: res.page, page_size: res.page_size, total: res.total, has_next: res.has_next }
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

function resetAndLoad() {
  page.value.page = 1
  load()
}

async function exportCsv() {
  try {
    const blob = await AuditService.exportCsv({ page: 1, page_size: 10000 })
    downloadBlob(blob, 'audit.csv')
    notifications.push('success', '已导出 audit.csv')
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

async function exportExceptions() {
  const cycleId = prompt('请输入要导出异常的周期 ID')
  if (!cycleId) return
  try {
    const blob = await AuditService.exportExceptionsCsv(cycleId)
    downloadBlob(blob, `exceptions-${cycleId}.csv`)
    notifications.push('success', '已导出异常 CSV')
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="审计与导出" description="只读审计日志, 支持 CSV 导出">
      <template #actions>
        <button class="btn" :disabled="loading" @click="load">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
        <button class="btn btn-primary" @click="exportCsv">导出审计 CSV</button>
        <button class="btn" @click="exportExceptions">导出异常 CSV</button>
      </template>
    </PageHeader>

    <div class="card">
      <div class="grid grid-3">
        <div class="field">
          <label>操作人 ID</label>
          <input class="input" v-model="filters.actor_id" @keyup.enter="resetAndLoad" />
        </div>
        <div class="field">
          <label>动作</label>
          <input class="input" v-model="filters.action" placeholder="例如 import/assign/resolve" @keyup.enter="resetAndLoad" />
        </div>
        <div class="field">
          <label>实体 ID</label>
          <input class="input" v-model="filters.entity_id" @keyup.enter="resetAndLoad" />
        </div>
      </div>
      <button class="btn btn-primary" @click="resetAndLoad">查询</button>
    </div>

    <div class="card">
      <table v-if="items.length" class="table">
        <thead>
          <tr>
            <th>时间</th>
            <th>操作人</th>
            <th>动作</th>
            <th>实体类型</th>
            <th>实体 ID</th>
            <th>详情</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in items" :key="e.id">
            <td class="text-sm">{{ e.created_at }}</td>
            <td>{{ e.actor_id }}</td>
            <td><span class="tag">{{ e.action }}</span></td>
            <td>{{ e.entity_type }}</td>
            <td class="text-sm">{{ e.entity_id }}</td>
            <td class="text-sm">
              <pre style="margin: 0; font-size: 11px">{{ JSON.stringify(e.detail || {}) }}</pre>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">暂无审计记录</div>
      <PaginationBar :page="page" @change="(p) => { page.page = p; load() }" />
    </div>
  </div>
</template>
