<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ExceptionsService } from '@/api/services'
import type { Exception } from '@/types'
import PageHeader from '@/components/PageHeader.vue'
import SeverityTag from '@/components/SeverityTag.vue'
import { useSessionStore } from '@/stores/session'
import { useNotificationsStore } from '@/stores/notifications'

const route = useRoute()
const router = useRouter()
const notifications = useNotificationsStore()
const session = useSessionStore()

const ex = ref<Exception | null>(null)
const loading = ref(false)
const assignForm = reactive({ assignee_id: '', note: '' })
const noteForm = reactive({ body: '', kind: 'comment' })
const actionForm = reactive({ note: '', reason: '' })

async function load() {
  loading.value = true
  try {
    ex.value = await ExceptionsService.get(route.params.id as string)
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  } finally {
    loading.value = false
  }
}

async function call(action: () => Promise<Exception>, successMsg: string) {
  try {
    ex.value = await action()
    notifications.push('success', successMsg)
  } catch (e: unknown) {
    notifications.push('error', (e as Error).message)
  }
}

function assign() {
  call(
    () => ExceptionsService.assign(ex.value!.id, assignForm.assignee_id, assignForm.note),
    '已分派',
  )
}

function claim() {
  if (!session.user) {
    notifications.push('info', '请先登录')
    return
  }
  call(() => ExceptionsService.claim(ex.value!.id, '认领'), '已认领')
}

function review() {
  call(() => ExceptionsService.review(ex.value!.id, actionForm.note), '已提交复核')
}

function resolve() {
  call(() => ExceptionsService.resolve(ex.value!.id, actionForm.note), '已解决')
}

function close() {
  call(() => ExceptionsService.close(ex.value!.id, actionForm.note), '已关闭')
}

function escalate() {
  if (!actionForm.reason) {
    notifications.push('info', '请填写升级原因')
    return
  }
  call(() => ExceptionsService.escalate(ex.value!.id, actionForm.reason), '已升级')
}

function rework() {
  if (!actionForm.note) {
    notifications.push('info', '请填写返工说明')
    return
  }
  call(() => ExceptionsService.rework(ex.value!.id, actionForm.note), '已返工')
}

function resubmit() {
  if (!actionForm.note) {
    notifications.push('info', '请填写重新提交说明')
    return
  }
  call(() => ExceptionsService.resubmit(ex.value!.id, actionForm.note), '已请求重新提交')
}

function addNote() {
  call(
    () => ExceptionsService.addNote(ex.value!.id, noteForm.body, noteForm.kind),
    '备注已追加',
  )
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader
      title="异常详情"
      :description="ex ? `${ex.title} (${ex.id})` : ''"
    >
      <template #actions>
        <button class="btn" :disabled="loading" @click="load">
          <span v-if="loading" class="spinner"></span>
          刷新
        </button>
        <button class="btn" @click="router.push({ name: 'exceptions' })">返回列表</button>
      </template>
    </PageHeader>

    <div v-if="ex" class="grid grid-2">
      <div class="card">
        <h3 class="text-lg" style="margin: 0 0 8px">基本信息</h3>
        <div class="flex flex-col gap-2 text-sm">
          <div><span class="muted">规则代码:</span> {{ ex.rule_code }}</div>
          <div><span class="muted">严重度:</span> <SeverityTag :severity="ex.severity" /></div>
          <div><span class="muted">状态:</span> <SeverityTag :status="ex.status" /></div>
          <div><span class="muted">版本:</span> {{ ex.version }}</div>
          <div><span class="muted">分派人:</span> {{ ex.assignee_id || '-' }}</div>
          <div><span class="muted">报告人:</span> {{ ex.reporter_id || '-' }}</div>
          <div><span class="muted">截止时间:</span> {{ ex.deadline_at || '-' }}</div>
          <div><span class="muted">解决时间:</span> {{ ex.resolved_at || '-' }}</div>
          <div><span class="muted">关闭时间:</span> {{ ex.closed_at || '-' }}</div>
          <div><span class="muted">命中原因:</span> {{ ex.hit_reason }}</div>
          <div><span class="muted">描述:</span> {{ ex.description }}</div>
        </div>
      </div>

      <div class="card">
        <h3 class="text-lg" style="margin: 0 0 8px">输入快照 (异常创建时锁定)</h3>
        <div class="flex flex-col gap-2 text-sm">
          <div><span class="muted">明细金额:</span> {{ ex.snapshot.entry_amount_cents }} 分</div>
          <div><span class="muted">币种:</span> {{ ex.snapshot.entry_currency }}</div>
          <div><span class="muted">发生时间:</span> {{ ex.snapshot.entry_occurred_at }}</div>
          <div><span class="muted">规则表达式:</span> {{ ex.snapshot.rule_expression }}</div>
          <div><span class="muted">规则严重度:</span> {{ ex.snapshot.rule_severity }}</div>
          <div><span class="muted">快照时间:</span> {{ ex.snapshot.snapshot_at }}</div>
          <div v-if="ex.snapshot.input_fields">
            <span class="muted">输入字段:</span>
            <pre style="margin: 4px 0 0; font-size: 12px; background: #fafbfc; padding: 8px">{{ JSON.stringify(ex.snapshot.input_fields, null, 2) }}</pre>
          </div>
        </div>
      </div>
    </div>

    <div v-if="ex" class="grid grid-2">
      <div class="card">
        <h3 class="text-lg" style="margin: 0 0 8px">状态变更</h3>
        <div class="field">
          <label>分派人 ID</label>
          <input class="input" v-model="assignForm.assignee_id" placeholder="例如: assignee" />
        </div>
        <div class="field">
          <label>分派备注</label>
          <input class="input" v-model="assignForm.note" />
        </div>
        <div class="flex gap-2" style="flex-wrap: wrap">
          <button class="btn btn-primary" @click="assign">分派</button>
          <button class="btn" @click="claim">认领</button>
          <button class="btn" @click="review">提交复核</button>
          <button class="btn" @click="resolve">解决</button>
          <button class="btn" @click="close">关闭</button>
          <button class="btn" @click="resubmit">请求重新提交</button>
          <button class="btn" @click="rework">返工</button>
          <button class="btn btn-danger" @click="escalate">升级</button>
        </div>
        <div class="field" style="margin-top: 12px">
          <label>升级原因 / 返工 / 重新提交说明</label>
          <input class="input" v-model="actionForm.reason" placeholder="升级原因" />
          <input class="input" v-model="actionForm.note" placeholder="通用备注 (返工/重新提交/复核/解决/关闭)" style="margin-top: 6px" />
        </div>
      </div>

      <div class="card">
        <h3 class="text-lg" style="margin: 0 0 8px">追加备注</h3>
        <div class="field">
          <label>类型</label>
          <select class="select" v-model="noteForm.kind">
            <option v-for="k in ['comment', 'assignment', 'claim', 'resubmit', 'review', 'escalation', 'rework']" :key="k" :value="k">
              {{ k }}
            </option>
          </select>
        </div>
        <div class="field">
          <label>内容</label>
          <textarea class="textarea" v-model="noteForm.body" rows="3"></textarea>
        </div>
        <button class="btn btn-primary" @click="addNote">追加</button>
      </div>
    </div>

    <div v-if="ex && ex.notes.length" class="card">
      <h3 class="text-lg" style="margin: 0 0 12px">备注历史</h3>
      <table class="table">
        <thead>
          <tr>
            <th>时间</th>
            <th>作者</th>
            <th>类型</th>
            <th>内容</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="n in ex.notes" :key="n.id">
            <td class="text-sm">{{ n.created_at }}</td>
            <td>{{ n.author_id }}</td>
            <td><span class="tag">{{ n.kind }}</span></td>
            <td>{{ n.body }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
