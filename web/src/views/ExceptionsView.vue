<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ExceptionsService } from '@/api/services'
import type { Exception, PageResult, ExceptionStatus, Severity } from '@/types'
import PageHeader from '@/components/PageHeader.vue'
import PaginationBar from '@/components/PaginationBar.vue'
import SeverityTag from '@/components/SeverityTag.vue'
import { useNotificationsStore } from '@/stores/notifications'

const notifications = useNotificationsStore()
const router = useRouter()

const items = ref<Exception[]>([])
const page = ref<PageResult>({ page: 1, page_size: 20, total: 0, has_next: false })
const loading = ref(false)

const filters = reactive({
  cycle_id: '',
  status: '' as '' | ExceptionStatus,
  severity: '' as '' | Severity,
  assignee_id: '',
  overdue_only: false,
})

async function load() {
  loading.value = true
  try {
    const res = await ExceptionsService.list({
      page: page.value.page,
      page_size: page.value.page_size,
      cycle_id: filters.cycle_id || undefined,
      status: filters.status || undefined,
      severity: filters.severity || undefined,
      assignee_id: filters.assignee_id || undefined,
      overdue_only: filters.overdue_only || undefined,
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

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="异常处置" description="支持分派 / 认领 / 复核 / 解决 / 升级 / 返工">
      <template #actions>
        <button class="btn" :disabled="loading" @click="load">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
      </template>
    </PageHeader>

    <div class="card">
      <div class="grid grid-3">
        <div class="field">
          <label>周期 ID</label>
          <input class="input" v-model="filters.cycle_id" @keyup.enter="resetAndLoad" />
        </div>
        <div class="field">
          <label>状态</label>
          <select class="select" v-model="filters.status" @change="resetAndLoad">
            <option value="">全部</option>
            <option v-for="s in ['pending', 'processing', 'review', 'resolved', 'closed', 'escalated']" :key="s" :value="s">
              {{ s }}
            </option>
          </select>
        </div>
        <div class="field">
          <label>严重度</label>
          <select class="select" v-model="filters.severity" @change="resetAndLoad">
            <option value="">全部</option>
            <option v-for="s in ['low', 'medium', 'high', 'critical']" :key="s" :value="s">
              {{ s }}
            </option>
          </select>
        </div>
        <div class="field">
          <label>分派人</label>
          <input class="input" v-model="filters.assignee_id" @keyup.enter="resetAndLoad" />
        </div>
        <div class="field">
          <label>仅超期</label>
          <input type="checkbox" v-model="filters.overdue_only" @change="resetAndLoad" />
        </div>
        <div class="field" style="justify-content: flex-end">
          <button class="btn btn-primary" @click="resetAndLoad">查询</button>
        </div>
      </div>
    </div>

    <div class="card">
      <table v-if="items.length" class="table">
        <thead>
          <tr>
            <th>异常 ID</th>
            <th>规则</th>
            <th>严重度</th>
            <th>状态</th>
            <th>分派人</th>
            <th>截止时间</th>
            <th>创建时间</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in items" :key="e.id">
            <td>
              <RouterLink :to="{ name: 'exception-detail', params: { id: e.id } }">{{ e.id.slice(0, 8) }}…</RouterLink>
            </td>
            <td>{{ e.rule_code }}</td>
            <td><SeverityTag :severity="e.severity" /></td>
            <td><SeverityTag :status="e.status" /></td>
            <td class="text-sm">{{ e.assignee_id || '-' }}</td>
            <td class="text-sm">{{ e.deadline_at || '-' }}</td>
            <td class="text-sm">{{ e.created_at }}</td>
            <td>
              <button class="btn-link btn" @click="router.push({ name: 'exception-detail', params: { id: e.id } })">
                处理
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">暂无异常</div>
      <PaginationBar :page="page" @change="(p) => { page.page = p; load() }" />
    </div>
  </div>
</template>
