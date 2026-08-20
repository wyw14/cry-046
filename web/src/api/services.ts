import { http } from './http'
import type {
  AnnualAccumulator,
  AuditEntry,
  Exception,
  FundingBatch,
  ListResponse,
  Party,
  Project,
  RuleVersion,
  SettlementCycle,
  SettlementEntry,
  Summary,
  UpsertSummary,
  WorkspaceView,
} from '@/types'

// === 项目 ===

export interface CreateProjectInput {
  code: string
  name: string
  sponsor: string
  annual_budget_cents: number
  start_year: number
  end_year: number
  metadata?: Record<string, string>
}

export interface ListParams {
  page?: number
  page_size?: number
  order_by?: string
  order?: 'asc' | 'desc'
  filter?: Record<string, string>
}

export const ProjectsService = {
  list(params: ListParams = {}) {
    return http.get<ListResponse<Project>>('/projects', { params: toQuery(params) }).then((r) => r.data)
  },
  get(id: string) {
    return http.get<Project>(`/projects/${id}`).then((r) => r.data)
  },
  create(input: CreateProjectInput) {
    return http.post<Project>('/projects', input).then((r) => r.data)
  },
  update(id: string, input: CreateProjectInput) {
    return http.patch<Project>(`/projects/${id}`, input).then((r) => r.data)
  },
}

// === 参与方 ===

export interface CreatePartyInput {
  name: string
  type: Party['type']
  contact: string
  metadata?: Record<string, string>
}

export const PartiesService = {
  list(params: ListParams = {}) {
    return http.get<ListResponse<Party>>('/parties', { params: toQuery(params) }).then((r) => r.data)
  },
  get(id: string) {
    return http.get<Party>(`/parties/${id}`).then((r) => r.data)
  },
  create(input: CreatePartyInput) {
    return http.post<Party>('/parties', input).then((r) => r.data)
  },
}

// === 资助批次 ===

export interface CreateBatchInput {
  project_id: string
  code: string
  donor_party_id: string
  implementer_party_id: string
  intermediary_party_id?: string
  total_amount_cents: number
  currency: string
  disbursed_at: string
  metadata?: Record<string, string>
}

export const BatchesService = {
  list(params: ListParams = {}) {
    return http.get<ListResponse<FundingBatch>>('/batches', { params: toQuery(params) }).then((r) => r.data)
  },
  get(id: string) {
    return http.get<FundingBatch>(`/batches/${id}`).then((r) => r.data)
  },
  create(input: CreateBatchInput) {
    return http.post<FundingBatch>('/batches', input).then((r) => r.data)
  },
}

// === 结算周期 ===

export interface CreateCycleInput {
  project_id: string
  funding_batch_id: string
  year: number
  quarter: 1 | 2 | 3 | 4
  start_date: string
  end_date: string
}

export const CyclesService = {
  list(params: ListParams = {}) {
    return http.get<ListResponse<SettlementCycle>>('/cycles', { params: toQuery(params) }).then((r) => r.data)
  },
  get(id: string) {
    return http.get<SettlementCycle>(`/cycles/${id}`).then((r) => r.data)
  },
  create(input: CreateCycleInput) {
    return http.post<SettlementCycle>('/cycles', input).then((r) => r.data)
  },
  close(id: string) {
    return http.post<SettlementCycle>(`/cycles/${id}/close`).then((r) => r.data)
  },
}

// === 规则版本 ===

export interface CreateRuleInput {
  code: string
  project_id: string
  description?: string
  rules: Array<{
    code: string
    description?: string
    severity: 'low' | 'medium' | 'high' | 'critical'
    category?: string
    expression: string
    deadline_hours?: number
  }>
}

export const RulesService = {
  list(params: ListParams = {}) {
    return http.get<ListResponse<RuleVersion>>('/rule-versions', { params: toQuery(params) }).then((r) => r.data)
  },
  get(id: string) {
    return http.get<RuleVersion>(`/rule-versions/${id}`).then((r) => r.data)
  },
  create(input: CreateRuleInput) {
    return http.post<RuleVersion>('/rule-versions', input).then((r) => r.data)
  },
  publish(id: string) {
    return http.post<RuleVersion>(`/rule-versions/${id}/publish`).then((r) => r.data)
  },
  archive(id: string) {
    return http.post<RuleVersion>(`/rule-versions/${id}/archive`).then((r) => r.data)
  },
}

// === 明细 ===

export interface ImportEntriesInput {
  batch_id: string
  cycle_id: string
  project_id: string
  entries: Array<{
    source_id: string
    source: 'import' | 'manual' | 'resubmit'
    payee_party_id: string
    payer_party_id: string
    amount_cents: number
    currency: string
    occurred_at: string
    metadata?: Record<string, string>
  }>
}

export interface ImportEntriesResponse {
  summary: UpsertSummary
  entries: SettlementEntry[]
}

export const EntriesService = {
  list(params: ListParams & { cycle_id?: string; batch_id?: string; project_id?: string; source?: string } = {}) {
    return http.get<ListResponse<SettlementEntry>>('/entries', { params: toQuery(params) }).then((r) => r.data)
  },
  get(id: string) {
    return http.get<SettlementEntry>(`/entries/${id}`).then((r) => r.data)
  },
  import(input: ImportEntriesInput) {
    return http.post<ImportEntriesResponse>('/entries/import', input).then((r) => r.data)
  },
}

// === 异常 ===

export interface ExceptionListParams extends ListParams {
  cycle_id?: string
  entry_id?: string
  status?: string
  severity?: string
  assignee_id?: string
  overdue_only?: boolean
}

export const ExceptionsService = {
  list(params: ExceptionListParams = {}) {
    return http.get<ListResponse<Exception>>('/exceptions', { params: toQuery(params) }).then((r) => r.data)
  },
  get(id: string) {
    return http.get<Exception>(`/exceptions/${id}`).then((r) => r.data)
  },
  assign(id: string, assignee_id: string, note?: string) {
    return http.post<Exception>(`/exceptions/${id}/assign`, { assignee_id, note }).then((r) => r.data)
  },
  claim(id: string, note?: string) {
    return http.post<Exception>(`/exceptions/${id}/claim`, { note }).then((r) => r.data)
  },
  resubmit(id: string, note: string) {
    return http.post<Exception>(`/exceptions/${id}/resubmit`, { note }).then((r) => r.data)
  },
  review(id: string, note?: string) {
    return http.post<Exception>(`/exceptions/${id}/review`, { note }).then((r) => r.data)
  },
  resolve(id: string, note?: string) {
    return http.post<Exception>(`/exceptions/${id}/resolve`, { note }).then((r) => r.data)
  },
  close(id: string, note?: string) {
    return http.post<Exception>(`/exceptions/${id}/close`, { note }).then((r) => r.data)
  },
  escalate(id: string, reason: string) {
    return http.post<Exception>(`/exceptions/${id}/escalate`, { reason }).then((r) => r.data)
  },
  rework(id: string, note: string) {
    return http.post<Exception>(`/exceptions/${id}/rework`, { note }).then((r) => r.data)
  },
  addNote(id: string, body: string, kind: string) {
    return http.post<Exception>(`/exceptions/${id}/notes`, { body, kind }).then((r) => r.data)
  },
}

// === 汇总 ===

export interface RecalcInput {
  cycle_id: string
  rule_version_id: string
  trigger_reason: string
}

export interface RecalcResponse {
  recalc_id: string
  summary: Summary
  previous: Summary | null
}

export const SummaryService = {
  recalculate(input: RecalcInput) {
    return http.post<RecalcResponse>('/summaries/recalculate', input).then((r) => r.data)
  },
  latest(cycleId: string) {
    return http.get<Summary>(`/summaries/cycles/${cycleId}/latest`).then((r) => r.data)
  },
  history(cycleId: string) {
    return http.get<{ items: Summary[] }>(`/summaries/cycles/${cycleId}/history`).then((r) => r.data)
  },
  recalcs(params: ListParams = {}) {
    return http.get<ListResponse<unknown>>('/summaries/recalcs', { params: toQuery(params) }).then((r) => r.data)
  },
  getAnnual(projectId: string, year: number) {
    return http.get<AnnualAccumulator>(`/annual/${projectId}/${year}`).then((r) => r.data)
  },
  adjustAnnual(projectId: string, year: number, delta_cents: number, reason: string) {
    return http
      .post<AnnualAccumulator>(`/annual/${projectId}/${year}/adjustments`, { delta_cents, reason })
      .then((r) => r.data)
  },
}

// === 工作台 ===

export const WorkspaceService = {
  get(assigneeId: string) {
    return http.get<WorkspaceView>(`/workspace/${assigneeId}`).then((r) => r.data)
  },
}

// === 审计 ===

export const AuditService = {
  list(params: ListParams & { actor_id?: string; action?: string; entity_id?: string } = {}) {
    return http.get<ListResponse<AuditEntry>>('/audit', { params: toQuery(params) }).then((r) => r.data)
  },
  exportCsv(params: ListParams = {}) {
    return http.get('/audit/export', { params: toQuery(params), responseType: 'blob' }).then((r) => r.data)
  },
  exportExceptionsCsv(cycleId: string) {
    return http
      .get(`/audit/exceptions/${cycleId}/export`, { responseType: 'blob' })
      .then((r) => r.data)
  },
}

// === 用户 ===

export interface CreateUserInput {
  username: string
  display_name?: string
  email?: string
  role: 'operator' | 'assignee' | 'reviewer' | 'admin'
  password_hash?: string
}

export const UsersService = {
  list(params: ListParams = {}) {
    return http
      .get<ListResponse<{ id: string; username: string; display_name: string; email: string; role: string }>>('/users', {
        params: toQuery(params),
      })
      .then((r) => r.data)
  },
  create(input: CreateUserInput) {
    return http.post('/users', input).then((r) => r.data)
  },
}

function toQuery(params: object): Record<string, string> {
  const out: Record<string, string> = {}
  for (const [k, v] of Object.entries(params as Record<string, unknown>)) {
    if (v === undefined || v === null || v === '') continue
    if (k === 'filter' && typeof v === 'object') {
      const filter = v as Record<string, string>
      for (const [fk, fv] of Object.entries(filter)) {
        if (fv === '' || fv === undefined || fv === null) continue
        out[`filter.${fk}`] = String(fv)
      }
      continue
    }
    out[k] = String(v)
  }
  return out
}
