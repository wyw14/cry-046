// 共享业务类型 — 与后端 DTO 一一对应.

export interface Project {
  id: string
  tenant_id: string
  code: string
  name: string
  sponsor: string
  annual_budget_cents: number
  start_year: number
  end_year: number
  metadata?: Record<string, string>
  created_at: string
  updated_at: string
}

export type PartyType = 'donor' | 'implementer' | 'beneficiary' | 'intermediary'

export interface Party {
  id: string
  tenant_id: string
  name: string
  type: PartyType
  contact: string
  metadata?: Record<string, string>
  created_at: string
  updated_at: string
}

export interface FundingBatch {
  id: string
  tenant_id: string
  project_id: string
  code: string
  donor_party_id: string
  implementer_party_id: string
  intermediary_party_id?: string
  total_amount_cents: number
  currency: string
  disbursed_at: string
  metadata?: Record<string, string>
  created_at: string
  updated_at: string
}

export interface SettlementCycle {
  id: string
  tenant_id: string
  project_id: string
  funding_batch_id: string
  year: number
  quarter: 1 | 2 | 3 | 4
  start_date: string
  end_date: string
  closed_at?: string
  created_at: string
  updated_at: string
}

export type Severity = 'low' | 'medium' | 'high' | 'critical'

export interface RuleDefinition {
  id: string
  code: string
  description: string
  severity: Severity
  category: string
  expression: string
  deadline_hours: number
}

export type RuleVersionStatus = 'draft' | 'published' | 'archived'

export interface RuleVersion {
  id: string
  tenant_id: string
  code: string
  project_id: string
  description: string
  rules: RuleDefinition[]
  status: RuleVersionStatus
  published_at?: string
  created_at: string
  updated_at: string
  version: number
}

export type EntrySource = 'import' | 'manual' | 'resubmit'

export interface SettlementEntry {
  id: string
  tenant_id: string
  cycle_id: string
  batch_id: string
  project_id: string
  source_id: string
  source: EntrySource
  payee_party_id: string
  payer_party_id: string
  amount_cents: number
  currency: string
  occurred_at: string
  source_fingerprint: string
  metadata?: Record<string, string>
  created_at: string
  updated_at: string
}

export type ExceptionStatus =
  | 'pending'
  | 'processing'
  | 'review'
  | 'resolved'
  | 'closed'
  | 'escalated'

export interface ExceptionNote {
  id: string
  author_id: string
  body: string
  created_at: string
  kind: string
}

export interface ExceptionAttachment {
  id: string
  original_name: string
  content_type: string
  size: number
  sha256: string
  uploaded_by: string
  created_at: string
}

export interface ExceptionSnapshot {
  entry_amount_cents: number
  entry_currency: string
  entry_occurred_at: string
  rule_expression: string
  rule_severity: string
  input_fields?: Record<string, string>
  snapshot_at: string
}

export interface Exception {
  id: string
  tenant_id: string
  cycle_id: string
  entry_id: string
  rule_version_id: string
  rule_code: string
  severity: Severity
  title: string
  description: string
  hit_reason: string
  status: ExceptionStatus
  assignee_id?: string
  reporter_id?: string
  deadline_at?: string
  resolved_at?: string
  closed_at?: string
  created_at: string
  updated_at: string
  version: number
  notes: ExceptionNote[]
  attachments: ExceptionAttachment[]
  snapshot: ExceptionSnapshot
}

export interface SummaryDiffBasis {
  previous_version: number
  previous_approved_cents: number
  delta_approved_cents: number
  trigger_reason: string
  trigger_exception_id?: string
  trigger_entry_id?: string
  trigger_rule_code?: string
}

export interface Summary {
  id: string
  tenant_id: string
  cycle_id: string
  rule_version_id: string
  computed_at: string
  total_entries: number
  total_amount_cents: number
  approved_amount_cents: number
  pending_amount_cents: number
  exception_count_by_status: Record<string, number>
  exception_count_by_severity: Record<string, number>
  diff_basis: SummaryDiffBasis
  version: number
}

export interface AuditEntry {
  id: string
  tenant_id: string
  actor_id: string
  action: string
  entity_type: string
  entity_id: string
  detail?: Record<string, string>
  created_at: string
}

export interface PageResult {
  page: number
  page_size: number
  total: number
  has_next: boolean
}

export interface ListResponse<T> {
  items: T[]
  page: number
  page_size: number
  total: number
  has_next: boolean
}

export interface ErrorEnvelope {
  code: string
  message: string
  fields?: { field: string; message: string }[]
  request_id: string
}

export interface UpsertSummary {
  created: number
  updated: number
  skipped: number
}

export interface AnnualAccumulator {
  project_id: string
  year: number
  budget_cents: number
  disbursed_cents: number
  available_cents: number
  overrun_cents: number
  adjustments?: Array<{
    id: string
    delta_cents: number
    reason: string
    triggered_by: string
    created_at: string
  }>
}

export interface WorkspaceView {
  assignee_id: string
  open: Exception[]
  overdue: Exception[]
  escalated: Exception[]
  recently_closed: Exception[]
}
