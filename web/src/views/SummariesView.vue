<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { SummaryService, CyclesService } from '@/api/services'
import type { Summary, SettlementCycle, AnnualAccumulator } from '@/types'
import PageHeader from '@/components/PageHeader.vue'
import { useNotificationsStore } from '@/stores/notifications'

const notifications = useNotificationsStore()

const cycles = ref<SettlementCycle[]>([])
const selectedCycleId = ref('')
const history = ref<Summary[]>([])
const latest = ref<Summary | null>(null)
const loading = ref(false)

const annualForm = reactive({
  project_id: '',
  year: new Date().getFullYear(),
  delta_cents: 0,
  reason: '',
})

const annual = ref<AnnualAccumulator | null>(null)

async function loadCycles() {
  try {
    const res = await CyclesService.list({ page: 1, page_size: 200 })
    cycles.value = res.items
    if (res.items.length && !selectedCycleId.value) {
      selectedCycleId.value = res.items[0].id
      await loadHistory()
    }
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

async function loadHistory() {
  if (!selectedCycleId.value) return
  loading.value = true
  try {
    const [hist, lat] = await Promise.all([
      SummaryService.history(selectedCycleId.value),
      SummaryService.latest(selectedCycleId.value).catch(() => null),
    ])
    history.value = hist.items || []
    latest.value = lat
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

async function loadAnnual() {
  if (!annualForm.project_id || !annualForm.year) return
  try {
    annual.value = await SummaryService.getAnnual(annualForm.project_id, annualForm.year)
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

async function applyAdjustment() {
  if (!annualForm.reason) {
    notifications.push('info', '请填写调整原因')
    return
  }
  try {
    annual.value = await SummaryService.adjustAnnual(
      annualForm.project_id,
      annualForm.year,
      annualForm.delta_cents,
      annualForm.reason,
    )
    notifications.push('success', '年度预算已调整')
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

const fmtCents = (n: number) => (n / 100).toFixed(2)

onMounted(loadCycles)
</script>

<template>
  <div>
    <PageHeader title="汇总与重算" description="查看历史汇总, 触发重算, 调整年度预算">
      <template #actions>
        <button class="btn" :disabled="loading" @click="loadHistory">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
      </template>
    </PageHeader>

    <div class="card">
      <div class="flex gap-3" style="align-items: flex-end; flex-wrap: wrap">
        <div class="field" style="flex: 1; min-width: 240px">
          <label>结算周期</label>
          <select class="select" v-model="selectedCycleId" @change="loadHistory">
            <option v-for="c in cycles" :key="c.id" :value="c.id">
              {{ c.year }} Q{{ c.quarter }} ({{ c.id.slice(0, 8) }}…)
            </option>
          </select>
        </div>
      </div>
    </div>

    <div v-if="latest" class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">最新汇总 (版本 {{ latest.version }})</h3>
      <div class="grid grid-3">
        <div class="stat-card">
          <div class="label">总金额</div>
          <div class="value">¥{{ fmtCents(latest.total_amount_cents) }}</div>
        </div>
        <div class="stat-card">
          <div class="label">已批准金额</div>
          <div class="value" style="color: var(--color-success)">¥{{ fmtCents(latest.approved_amount_cents) }}</div>
        </div>
        <div class="stat-card">
          <div class="label">待处理金额</div>
          <div class="value" style="color: var(--color-warning)">¥{{ fmtCents(latest.pending_amount_cents) }}</div>
        </div>
      </div>
      <div v-if="latest.diff_basis" class="muted text-sm" style="margin-top: 12px">
        差量: {{ latest.diff_basis.delta_approved_cents }} 分 · 触发原因: {{ latest.diff_basis.trigger_reason }}
      </div>
    </div>

    <div class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">汇总历史 (保留快照)</h3>
      <table v-if="history.length" class="table">
        <thead>
          <tr>
            <th>版本</th>
            <th>计算时间</th>
            <th>总金额</th>
            <th>已批准</th>
            <th>待处理</th>
            <th>差量</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in history" :key="s.id">
            <td>{{ s.version }}</td>
            <td class="text-sm">{{ s.computed_at }}</td>
            <td>¥{{ fmtCents(s.total_amount_cents) }}</td>
            <td>¥{{ fmtCents(s.approved_amount_cents) }}</td>
            <td>¥{{ fmtCents(s.pending_amount_cents) }}</td>
            <td>{{ s.diff_basis?.delta_approved_cents ?? 0 }} 分</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">暂无汇总历史</div>
    </div>

    <div class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">年度预算调整</h3>
      <div class="grid grid-2">
        <div class="field">
          <label>项目 ID</label>
          <input class="input" v-model="annualForm.project_id" @change="loadAnnual" />
        </div>
        <div class="field">
          <label>年度</label>
          <input class="input" type="number" v-model.number="annualForm.year" @change="loadAnnual" />
        </div>
        <div class="field">
          <label>调整金额 (分, 可负)</label>
          <input class="input" type="number" v-model.number="annualForm.delta_cents" />
        </div>
        <div class="field">
          <label>调整原因</label>
          <input class="input" v-model="annualForm.reason" />
        </div>
      </div>
      <button class="btn btn-primary" @click="applyAdjustment">应用调整</button>

      <div v-if="annual" class="grid grid-3" style="margin-top: 16px">
        <div class="stat-card">
          <div class="label">预算</div>
          <div class="value">¥{{ fmtCents(annual.budget_cents) }}</div>
        </div>
        <div class="stat-card">
          <div class="label">已拨付</div>
          <div class="value">¥{{ fmtCents(annual.disbursed_cents) }}</div>
        </div>
        <div class="stat-card">
          <div class="label">剩余 / 超支</div>
          <div
            class="value"
            :style="{ color: annual.overrun_cents > 0 ? 'var(--color-danger)' : 'var(--color-success)' }"
          >
            ¥{{ fmtCents(annual.available_cents) }}
          </div>
        </div>
      </div>
      <div v-if="annual && annual.adjustments && annual.adjustments.length" style="margin-top: 16px">
        <h4 style="margin: 0 0 8px">调整记录</h4>
        <table class="table">
          <thead>
            <tr>
              <th>时间</th>
              <th>分差 (分)</th>
              <th>原因</th>
              <th>操作人</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in annual.adjustments" :key="a.id">
              <td class="text-sm">{{ a.created_at }}</td>
              <td :style="{ color: a.delta_cents < 0 ? 'var(--color-danger)' : 'inherit' }">
                {{ a.delta_cents }}
              </td>
              <td>{{ a.reason }}</td>
              <td class="text-sm">{{ a.triggered_by }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
