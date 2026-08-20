<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  CyclesService,
  EntriesService,
  RulesService,
  ExceptionsService,
  SummaryService,
} from '@/api/services'
import type {
  SettlementCycle,
  SettlementEntry,
  Exception,
  RuleVersion,
  Summary,
} from '@/types'
import PageHeader from '@/components/PageHeader.vue'
import SeverityTag from '@/components/SeverityTag.vue'
import { useNotificationsStore } from '@/stores/notifications'

const route = useRoute()
const router = useRouter()
const notifications = useNotificationsStore()

const cycle = ref<SettlementCycle | null>(null)
const entries = ref<SettlementEntry[]>([])
const exceptions = ref<Exception[]>([])
const ruleVersions = ref<RuleVersion[]>([])
const summary = ref<Summary | null>(null)
const loading = ref(false)

const importForm = reactive({
  batch_id: '',
  project_id: '',
  rows: [
    {
      source_id: 'S-001',
      source: 'import' as 'import' | 'manual' | 'resubmit',
      payee_party_id: '',
      payer_party_id: '',
      amount_cents: 0,
      currency: 'CNY',
      occurred_at: new Date().toISOString(),
      metadata_note: '',
    },
  ],
})

const showImport = ref(false)
const selectedRuleVersionId = ref('')
const recalcReason = ref('手动触发重算')

async function load() {
  loading.value = true
  const id = route.params.id as string
  try {
    cycle.value = await CyclesService.get(id)
    const [ents, excs, rvList, summ] = await Promise.all([
      EntriesService.list({ cycle_id: id, page: 1, page_size: 200 }),
      ExceptionsService.list({ cycle_id: id, page: 1, page_size: 200 }),
      RulesService.list({ filter: { project_id: cycle.value.project_id } }),
      SummaryService.latest(id).catch(() => null),
    ])
    entries.value = ents.items
    exceptions.value = excs.items
    ruleVersions.value = rvList.items
    summary.value = summ
    if (rvList.items.length && !selectedRuleVersionId.value) {
      const published = rvList.items.find((rv) => rv.status === 'published')
      selectedRuleVersionId.value = (published || rvList.items[0]).id
    }
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

const fmtCents = (n: number) => (n / 100).toFixed(2)

const exceptionByEntry = computed(() => {
  const map = new Map<string, Exception[]>()
  for (const ex of exceptions.value) {
    const arr = map.get(ex.entry_id) || []
    arr.push(ex)
    map.set(ex.entry_id, arr)
  }
  return map
})

async function doImport() {
  if (!cycle.value) return
  try {
    const payload = {
      batch_id: importForm.batch_id || cycle.value.funding_batch_id,
      cycle_id: cycle.value.id,
      project_id: importForm.project_id || cycle.value.project_id,
      entries: importForm.rows.map((r) => ({
        source_id: r.source_id,
        source: r.source,
        payee_party_id: r.payee_party_id,
        payer_party_id: r.payer_party_id,
        amount_cents: r.amount_cents,
        currency: r.currency,
        occurred_at: r.occurred_at,
        metadata: r.metadata_note ? { note: r.metadata_note } : undefined,
      })),
    }
    const res = await EntriesService.import(payload)
    notifications.push('success', `导入完成: 新增 ${res.summary.created}, 更新 ${res.summary.updated}, 跳过 ${res.summary.skipped}`)
    showImport.value = false
    await load()
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

function addRow() {
  importForm.rows.push({
    source_id: 'S-' + (importForm.rows.length + 1).toString().padStart(3, '0'),
    source: 'import',
    payee_party_id: '',
    payer_party_id: '',
    amount_cents: 0,
    currency: 'CNY',
    occurred_at: new Date().toISOString(),
    metadata_note: '',
  })
}

function removeRow(i: number) {
  if (importForm.rows.length > 1) importForm.rows.splice(i, 1)
}

async function recalculate() {
  if (!cycle.value || !selectedRuleVersionId.value) {
    notifications.push('info', '请先选择规则版本')
    return
  }
  try {
    const res = await SummaryService.recalculate({
      cycle_id: cycle.value.id,
      rule_version_id: selectedRuleVersionId.value,
      trigger_reason: recalcReason.value,
    })
    summary.value = res.summary
    notifications.push('success', `重算完成: 已批准金额 ¥${fmtCents(res.summary.approved_amount_cents)}`)
    await load()
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader
      :title="`结算周期详情`"
      :description="cycle ? `${cycle.year} Q${cycle.quarter} (${cycle.id})` : ''"
    >
      <template #actions>
        <button class="btn" :disabled="loading" @click="load">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
        <button class="btn" @click="router.push({ name: 'projects' })">返回</button>
      </template>
    </PageHeader>

    <div v-if="cycle" class="grid grid-3">
      <div class="stat-card">
        <div class="label">结算明细</div>
        <div class="value">{{ entries.length }}</div>
      </div>
      <div class="stat-card">
        <div class="label">异常数</div>
        <div class="value" style="color: var(--color-warning)">{{ exceptions.length }}</div>
      </div>
      <div class="stat-card">
        <div class="label">已批准金额</div>
        <div class="value">¥{{ summary ? fmtCents(summary.approved_amount_cents) : '-' }}</div>
      </div>
    </div>

    <div v-if="summary" class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">最新汇总 (版本 {{ summary.version }})</h3>
      <div class="grid grid-3">
        <div><span class="muted">总金额:</span> ¥{{ fmtCents(summary.total_amount_cents) }}</div>
        <div><span class="muted">已批准:</span> ¥{{ fmtCents(summary.approved_amount_cents) }}</div>
        <div><span class="muted">待处理:</span> ¥{{ fmtCents(summary.pending_amount_cents) }}</div>
        <div><span class="muted">计算时间:</span> {{ summary.computed_at }}</div>
        <div><span class="muted">差量:</span> {{ summary.diff_basis.delta_approved_cents }} 分</div>
        <div><span class="muted">触发原因:</span> {{ summary.diff_basis.trigger_reason }}</div>
      </div>
    </div>

    <div class="card">
      <div class="flex items-center justify-between" style="margin-bottom: 12px">
        <h3 class="text-lg" style="margin: 0">结算明细</h3>
        <button class="btn btn-primary" @click="showImport = !showImport">导入明细</button>
      </div>
      <div v-if="showImport" class="card" style="background: #fafbfc; margin-bottom: 12px">
        <h4 style="margin: 0 0 8px">批量导入</h4>
        <div v-for="(r, i) in importForm.rows" :key="i" class="flex gap-2" style="margin-bottom: 8px">
          <input class="input" v-model="r.source_id" placeholder="source_id" style="flex: 1" />
          <select class="select" v-model="r.source" style="width: 100px">
            <option value="import">import</option>
            <option value="manual">manual</option>
            <option value="resubmit">resubmit</option>
          </select>
          <input class="input" v-model="r.payer_party_id" placeholder="付款方" style="flex: 1" />
          <input class="input" v-model="r.payee_party_id" placeholder="收款方" style="flex: 1" />
          <input class="input" type="number" v-model.number="r.amount_cents" placeholder="金额(分)" style="width: 100px" />
          <input class="input" v-model="r.currency" placeholder="币种" style="width: 60px" />
          <input class="input" type="datetime-local" v-model="r.occurred_at" style="width: 200px" />
          <button class="btn btn-danger" @click="removeRow(i)">删除</button>
        </div>
        <div class="flex gap-2">
          <button class="btn" @click="addRow">添加一行</button>
          <button class="btn btn-primary" @click="doImport">执行导入</button>
          <button class="btn" @click="showImport = false">取消</button>
        </div>
      </div>
      <table v-if="entries.length" class="table">
        <thead>
          <tr>
            <th>source_id</th>
            <th>付款方</th>
            <th>收款方</th>
            <th>金额</th>
            <th>币种</th>
            <th>发生时间</th>
            <th>异常</th>
            <th>指纹</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in entries" :key="e.id">
            <td>{{ e.source_id }}</td>
            <td class="text-sm">{{ e.payer_party_id }}</td>
            <td class="text-sm">{{ e.payee_party_id }}</td>
            <td>¥{{ fmtCents(e.amount_cents) }}</td>
            <td>{{ e.currency }}</td>
            <td class="text-sm">{{ e.occurred_at }}</td>
            <td>
              <SeverityTag
                v-if="exceptionByEntry.get(e.id)?.length"
                :severity="exceptionByEntry.get(e.id)![0].severity"
              />
              <span v-else class="muted">-</span>
            </td>
            <td class="text-sm muted" :title="e.source_fingerprint">{{ e.source_fingerprint.slice(0, 8) }}…</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">暂无明细</div>
    </div>

    <div class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">异常列表</h3>
      <table v-if="exceptions.length" class="table">
        <thead>
          <tr>
            <th>异常 ID</th>
            <th>规则</th>
            <th>严重度</th>
            <th>状态</th>
            <th>分派人</th>
            <th>截止时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in exceptions" :key="e.id">
            <td>
              <RouterLink :to="{ name: 'exception-detail', params: { id: e.id } }">{{ e.id.slice(0, 8) }}…</RouterLink>
            </td>
            <td>{{ e.rule_code }}</td>
            <td><SeverityTag :severity="e.severity" /></td>
            <td><SeverityTag :status="e.status" /></td>
            <td class="text-sm">{{ e.assignee_id || '-' }}</td>
            <td class="text-sm">{{ e.deadline_at || '-' }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">暂无异常</div>
    </div>

    <div class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">重新计算汇总</h3>
      <div class="field">
        <label>规则版本</label>
        <select class="select" v-model="selectedRuleVersionId">
          <option v-for="rv in ruleVersions" :key="rv.id" :value="rv.id">
            {{ rv.code }} ({{ rv.status }})
          </option>
        </select>
      </div>
      <div class="field">
        <label>触发原因</label>
        <input class="input" v-model="recalcReason" />
      </div>
      <button class="btn btn-primary" @click="recalculate">触发重算</button>
    </div>
  </div>
</template>
