<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ProjectsService, PartiesService, BatchesService, CyclesService } from '@/api/services'
import type { Project, Party, FundingBatch, SettlementCycle, PageResult } from '@/types'
import PageHeader from '@/components/PageHeader.vue'
import PaginationBar from '@/components/PaginationBar.vue'
import SeverityTag from '@/components/SeverityTag.vue'
import { useNotificationsStore } from '@/stores/notifications'

const notifications = useNotificationsStore()

type Tab = 'projects' | 'parties' | 'batches' | 'cycles'
const tab = ref<Tab>('projects')

const projects = ref<Project[]>([])
const parties = ref<Party[]>([])
const batches = ref<FundingBatch[]>([])
const cycles = ref<SettlementCycle[]>([])
const page = ref<PageResult>({ page: 1, page_size: 20, total: 0, has_next: false })
const loading = ref(false)
const showCreate = ref(false)

const form = reactive({
  code: '',
  name: '',
  sponsor: '示例资助方',
  annual_budget_cents: 10000000,
  start_year: 2026,
  end_year: 2027,
})

async function loadProjects() {
  loading.value = true
  try {
    const res = await ProjectsService.list({ page: page.value.page, page_size: page.value.page_size })
    projects.value = res.items
    page.value = { page: res.page, page_size: res.page_size, total: res.total, has_next: res.has_next }
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadParties() {
  loading.value = true
  try {
    const res = await PartiesService.list({ page: 1, page_size: 100 })
    parties.value = res.items
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadBatches() {
  loading.value = true
  try {
    const res = await BatchesService.list({ page: 1, page_size: 100 })
    batches.value = res.items
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadCycles() {
  loading.value = true
  try {
    const res = await CyclesService.list({ page: 1, page_size: 100 })
    cycles.value = res.items
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

async function createProject() {
  try {
    await ProjectsService.create({ ...form })
    notifications.push('success', '项目已创建')
    showCreate.value = false
    await loadProjects()
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

async function switchTab(t: Tab) {
  tab.value = t
  if (t === 'projects') await loadProjects()
  if (t === 'parties') await loadParties()
  if (t === 'batches') await loadBatches()
  if (t === 'cycles') await loadCycles()
}

onMounted(loadProjects)

function fmtCents(n: number): string {
  return (n / 100).toFixed(2)
}

const visibleItems = computed(() => {
  if (tab.value === 'projects') return projects.value
  if (tab.value === 'parties') return parties.value
  if (tab.value === 'batches') return batches.value
  return cycles.value
})
</script>

<template>
  <div>
    <PageHeader title="项目 / 参与方 / 批次 / 周期" description="项目元数据与结算周期管理">
      <template #actions>
        <button class="btn" :disabled="loading" @click="switchTab(tab)">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
        <button v-if="tab === 'projects'" class="btn btn-primary" @click="showCreate = !showCreate">
          新建项目
        </button>
      </template>
    </PageHeader>

    <div class="flex gap-2" style="margin-bottom: 12px">
      <button
        v-for="t in (['projects', 'parties', 'batches', 'cycles'] as Tab[])"
        :key="t"
        :class="['btn', tab === t ? 'btn-primary' : '']"
        @click="switchTab(t)"
      >
        {{ { projects: '项目', parties: '参与方', batches: '批次', cycles: '周期' }[t] }}
      </button>
    </div>

    <div v-if="showCreate && tab === 'projects'" class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">新建项目</h3>
      <div class="grid grid-2">
        <div class="field">
          <label>项目代码</label>
          <input class="input" v-model="form.code" placeholder="WS-2026-01" />
        </div>
        <div class="field">
          <label>项目名称</label>
          <input class="input" v-model="form.name" />
        </div>
        <div class="field">
          <label>资助方</label>
          <input class="input" v-model="form.sponsor" />
        </div>
        <div class="field">
          <label>年度预算 (分)</label>
          <input class="input" type="number" v-model.number="form.annual_budget_cents" />
        </div>
        <div class="field">
          <label>开始年度</label>
          <input class="input" type="number" v-model.number="form.start_year" />
        </div>
        <div class="field">
          <label>结束年度</label>
          <input class="input" type="number" v-model.number="form.end_year" />
        </div>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-primary" @click="createProject">创建</button>
        <button class="btn" @click="showCreate = false">取消</button>
      </div>
    </div>

    <div class="card">
      <table v-if="tab === 'projects'" class="table">
        <thead>
          <tr>
            <th>代码</th>
            <th>名称</th>
            <th>资助方</th>
            <th>年度预算</th>
            <th>年度范围</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in projects" :key="p.id">
            <td>{{ p.code }}</td>
            <td>{{ p.name }}</td>
            <td>{{ p.sponsor }}</td>
            <td>¥{{ fmtCents(p.annual_budget_cents) }}</td>
            <td>{{ p.start_year }} - {{ p.end_year }}</td>
          </tr>
        </tbody>
      </table>

      <table v-else-if="tab === 'parties'" class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>联系方式 (脱敏)</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in parties" :key="p.id">
            <td>{{ p.name }}</td>
            <td><SeverityTag :severity="undefined" /></td>
            <td>{{ p.contact }}</td>
          </tr>
        </tbody>
      </table>

      <table v-else-if="tab === 'batches'" class="table">
        <thead>
          <tr>
            <th>代码</th>
            <th>项目 ID</th>
            <th>总金额</th>
            <th>币种</th>
            <th>拨付时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in batches" :key="b.id">
            <td>{{ b.code }}</td>
            <td class="text-sm">{{ b.project_id }}</td>
            <td>¥{{ fmtCents(b.total_amount_cents) }}</td>
            <td>{{ b.currency }}</td>
            <td class="text-sm">{{ b.disbursed_at }}</td>
          </tr>
        </tbody>
      </table>

      <table v-else class="table">
        <thead>
          <tr>
            <th>年度</th>
            <th>季度</th>
            <th>项目 ID</th>
            <th>开始日期</th>
            <th>结束日期</th>
            <th>状态</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in cycles" :key="c.id">
            <td>{{ c.year }}</td>
            <td>Q{{ c.quarter }}</td>
            <td class="text-sm">{{ c.project_id }}</td>
            <td class="text-sm">{{ c.start_date }}</td>
            <td class="text-sm">{{ c.end_date }}</td>
            <td>
              <span v-if="c.closed_at" class="tag tag-closed">已关闭</span>
              <span v-else class="tag tag-pending">进行中</span>
            </td>
            <td>
              <RouterLink :to="{ name: 'cycle-detail', params: { id: c.id } }">查看</RouterLink>
            </td>
          </tr>
        </tbody>
      </table>

      <PaginationBar
        v-if="tab === 'projects'"
        :page="page"
        @change="(p) => { page.page = p; loadProjects() }"
      />
    </div>
  </div>
</template>
