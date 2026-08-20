<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useSessionStore } from '@/stores/session'
import { WorkspaceService } from '@/api/services'
import type { WorkspaceView } from '@/types'
import PageHeader from '@/components/PageHeader.vue'
import SeverityTag from '@/components/SeverityTag.vue'
import { useNotificationsStore } from '@/stores/notifications'

const session = useSessionStore()
const notifications = useNotificationsStore()
const view = ref<WorkspaceView | null>(null)
const loading = ref(false)
const error = ref('')

const assigneeId = computed(() => session.user?.username || 'admin')

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    view.value = await WorkspaceService.get(assigneeId.value)
  } catch (e: unknown) {
    error.value = (e as Error).message
    notifications.push('error', error.value)
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>

<template>
  <div>
    <PageHeader
      title="工作台"
      :description="`处理人: ${assigneeId} (${session.user?.role || '-'})`"
    >
      <template #actions>
        <button class="btn" :disabled="loading" @click="refresh">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
      </template>
    </PageHeader>

    <div v-if="error" class="card" style="color: var(--color-danger)">{{ error }}</div>

    <div v-if="view" class="grid grid-3">
      <div class="stat-card">
        <div class="label">待处理</div>
        <div class="value">{{ view.open.length }}</div>
      </div>
      <div class="stat-card">
        <div class="label">超期</div>
        <div class="value" style="color: var(--color-danger)">{{ view.overdue.length }}</div>
      </div>
      <div class="stat-card">
        <div class="label">已升级</div>
        <div class="value" style="color: var(--color-warning)">{{ view.escalated.length }}</div>
      </div>
    </div>

    <div v-if="view" class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">待处理异常</h3>
      <table v-if="view.open.length" class="table">
        <thead>
          <tr>
            <th>异常 ID</th>
            <th>规则</th>
            <th>严重度</th>
            <th>状态</th>
            <th>截止时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in view.open" :key="e.id">
            <td>
              <RouterLink :to="{ name: 'exception-detail', params: { id: e.id } }">{{ e.id }}</RouterLink>
            </td>
            <td>{{ e.rule_code }}</td>
            <td><SeverityTag :severity="e.severity" /></td>
            <td><SeverityTag :status="e.status" /></td>
            <td class="text-sm muted">{{ e.deadline_at || '-' }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">暂无待处理异常</div>
    </div>

    <div v-if="view && view.overdue.length" class="card">
      <h3 class="text-lg" style="margin: 0 0 12px; color: var(--color-danger)">超期异常</h3>
      <table class="table">
        <thead>
          <tr>
            <th>异常 ID</th>
            <th>规则</th>
            <th>严重度</th>
            <th>截止时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in view.overdue" :key="e.id">
            <td>
              <RouterLink :to="{ name: 'exception-detail', params: { id: e.id } }">{{ e.id }}</RouterLink>
            </td>
            <td>{{ e.rule_code }}</td>
            <td><SeverityTag :severity="e.severity" /></td>
            <td class="text-sm">{{ e.deadline_at }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
