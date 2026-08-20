package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// RuleVersionRepo is a pgx-backed RuleVersionRepo.
type RuleVersionRepo struct{ pool *pgxpool.Pool }

func (r *RuleVersionRepo) Create(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error) {
	if err := rv.Validate(); err != nil {
		return domain.RuleVersion{}, err
	}
	rulesJSON, err := json.Marshal(rv.Rules)
	if err != nil {
		return domain.RuleVersion{}, fmt.Errorf("marshal rules: %w", err)
	}
	const q = `INSERT INTO rule_versions (id, tenant_id, code, project_id, description, rules, status,
		published_at, created_at, updated_at, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, code) DO NOTHING`
	_, err = r.pool.Exec(ctx, q, rv.ID, rv.TenantID, rv.Code, rv.ProjectID, rv.Description,
		rulesJSON, rv.Status, nullableTime(rv.PublishedAt), rv.CreatedAt, rv.UpdatedAt, rv.Version)
	if err != nil {
		return domain.RuleVersion{}, translateError(err)
	}
	return rv, nil
}

func (r *RuleVersionRepo) Get(ctx context.Context, tenantID, id string) (domain.RuleVersion, error) {
	const q = `SELECT id, tenant_id, code, project_id, description, rules, status,
		published_at, created_at, updated_at, version
		FROM rule_versions WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var rv domain.RuleVersion
	var publishedAt *time.Time
	var rulesJSON []byte
	err := row.Scan(&rv.ID, &rv.TenantID, &rv.Code, &rv.ProjectID, &rv.Description,
		&rulesJSON, &rv.Status, &publishedAt, &rv.CreatedAt, &rv.UpdatedAt, &rv.Version)
	if err != nil {
		return domain.RuleVersion{}, translateError(err)
	}
	if publishedAt != nil {
		rv.PublishedAt = *publishedAt
	}
	if err := json.Unmarshal(rulesJSON, &rv.Rules); err != nil {
		return domain.RuleVersion{}, fmt.Errorf("unmarshal rules: %w", err)
	}
	return rv, nil
}

func (r *RuleVersionRepo) GetByCode(ctx context.Context, tenantID, code string) (domain.RuleVersion, error) {
	const q = `SELECT id, tenant_id, code, project_id, description, rules, status,
		published_at, created_at, updated_at, version
		FROM rule_versions WHERE tenant_id=$1 AND code=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, code)
	var rv domain.RuleVersion
	var publishedAt *time.Time
	var rulesJSON []byte
	err := row.Scan(&rv.ID, &rv.TenantID, &rv.Code, &rv.ProjectID, &rv.Description,
		&rulesJSON, &rv.Status, &publishedAt, &rv.CreatedAt, &rv.UpdatedAt, &rv.Version)
	if err != nil {
		return domain.RuleVersion{}, translateError(err)
	}
	if publishedAt != nil {
		rv.PublishedAt = *publishedAt
	}
	if err := json.Unmarshal(rulesJSON, &rv.Rules); err != nil {
		return domain.RuleVersion{}, fmt.Errorf("unmarshal rules: %w", err)
	}
	return rv, nil
}

func (r *RuleVersionRepo) List(ctx context.Context, q application.ListQuery) ([]domain.RuleVersion, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	for k, v := range q.Filters {
		col, ok := ruleVersionFilterColumn(k)
		if !ok {
			return nil, application.PageResult{}, domain.NewErrf(domain.CodeInvalidArgument, "unknown filter %s", k).WithField("filter")
		}
		where = append(where, fmt.Sprintf("%s=$%d", col, i))
		args = append(args, v)
		i++
	}
	orderBy := "code"
	if q.OrderBy != "" {
		if col, ok := ruleVersionFilterColumn(q.OrderBy); ok {
			orderBy = col
		}
	}
	dir := "ASC"
	if q.OrderDesc {
		dir = "DESC"
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, tenant_id, code, project_id, description, rules, status,
		published_at, created_at, updated_at, version
		FROM rule_versions WHERE %s ORDER BY %s %s, id ASC LIMIT %d OFFSET %d`,
		joinStrings(where, " AND "), orderBy, dir, pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.RuleVersion, 0, pageSize)
	for rows.Next() {
		var rv domain.RuleVersion
		var publishedAt *time.Time
		var rulesJSON []byte
		if err := rows.Scan(&rv.ID, &rv.TenantID, &rv.Code, &rv.ProjectID, &rv.Description,
			&rulesJSON, &rv.Status, &publishedAt, &rv.CreatedAt, &rv.UpdatedAt, &rv.Version); err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		if publishedAt != nil {
			rv.PublishedAt = *publishedAt
		}
		if err := json.Unmarshal(rulesJSON, &rv.Rules); err != nil {
			return nil, application.PageResult{}, fmt.Errorf("unmarshal rules: %w", err)
		}
		out = append(out, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM rule_versions WHERE %s`, joinStrings(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func (r *RuleVersionRepo) Update(ctx context.Context, rv domain.RuleVersion) (domain.RuleVersion, error) {
	if err := rv.Validate(); err != nil {
		return domain.RuleVersion{}, err
	}
	rulesJSON, err := json.Marshal(rv.Rules)
	if err != nil {
		return domain.RuleVersion{}, fmt.Errorf("marshal rules: %w", err)
	}
	const q = `UPDATE rule_versions SET description=$1, rules=$2, status=$3, published_at=$4,
		updated_at=$5, version=$6 WHERE tenant_id=$7 AND id=$8`
	_, err = r.pool.Exec(ctx, q, rv.Description, rulesJSON, rv.Status, nullableTime(rv.PublishedAt),
		rv.UpdatedAt, rv.Version, rv.TenantID, rv.ID)
	if err != nil {
		return domain.RuleVersion{}, translateError(err)
	}
	return rv, nil
}

func ruleVersionFilterColumn(k string) (string, bool) {
	switch k {
	case "code":
		return "code", true
	case "project_id":
		return "project_id", true
	case "status":
		return "status", true
	}
	return "", false
}

// EntryRepo is a pgx-backed EntryRepo.
type EntryRepo struct{ pool *pgxpool.Pool }

func (r *EntryRepo) UpsertBatch(ctx context.Context, entries []domain.SettlementEntry) (application.UpsertSummary, error) {
	var s application.UpsertSummary
	for _, e := range entries {
		if err := e.Validate(); err != nil {
			return s, err
		}
		metaJSON, err := json.Marshal(e.Metadata)
		if err != nil {
			return s, fmt.Errorf("marshal metadata: %w", err)
		}
		const q = `INSERT INTO settlement_entries (id, tenant_id, cycle_id, batch_id, project_id, source_id,
			source, payee_party_id, payer_party_id, amount_cents, currency, occurred_at,
			source_fingerprint, metadata, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			ON CONFLICT (source_fingerprint) DO UPDATE SET
				amount_cents=EXCLUDED.amount_cents,
				currency=EXCLUDED.currency,
				occurred_at=EXCLUDED.occurred_at,
				metadata=EXCLUDED.metadata,
				updated_at=EXCLUDED.updated_at
			RETURNING (xmax = 0) AS inserted`
		var inserted bool
		err = r.pool.QueryRow(ctx, q, e.ID, e.TenantID, e.CycleID, e.BatchID, e.ProjectID, e.SourceID,
			e.Source, e.PayeePartyID, e.PayerPartyID, e.Amount, e.Currency, e.OccurredAt,
			e.SourceFingerprint, metaJSON, e.CreatedAt, e.UpdatedAt).Scan(&inserted)
		if err != nil {
			return s, translateError(err)
		}
		if inserted {
			s.Created++
		} else {
			s.Updated++
		}
	}
	return s, nil
}

func (r *EntryRepo) Get(ctx context.Context, tenantID, id string) (domain.SettlementEntry, error) {
	const q = `SELECT id, tenant_id, cycle_id, batch_id, project_id, source_id, source,
		payee_party_id, payer_party_id, amount_cents, currency, occurred_at,
		source_fingerprint, metadata, created_at, updated_at
		FROM settlement_entries WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var e domain.SettlementEntry
	var metaJSON []byte
	err := row.Scan(&e.ID, &e.TenantID, &e.CycleID, &e.BatchID, &e.ProjectID, &e.SourceID, &e.Source,
		&e.PayeePartyID, &e.PayerPartyID, &e.Amount, &e.Currency, &e.OccurredAt,
		&e.SourceFingerprint, &metaJSON, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.SettlementEntry{}, translateError(err)
	}
	if metaJSON != nil {
		_ = json.Unmarshal(metaJSON, &e.Metadata)
	}
	return e, nil
}

func (r *EntryRepo) List(ctx context.Context, q application.EntryListQuery) ([]domain.SettlementEntry, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	if q.CycleID != "" {
		where = append(where, fmt.Sprintf("cycle_id=$%d", i))
		args = append(args, q.CycleID)
		i++
	}
	if q.BatchID != "" {
		where = append(where, fmt.Sprintf("batch_id=$%d", i))
		args = append(args, q.BatchID)
		i++
	}
	if q.ProjectID != "" {
		where = append(where, fmt.Sprintf("project_id=$%d", i))
		args = append(args, q.ProjectID)
		i++
	}
	if q.Source != "" {
		where = append(where, fmt.Sprintf("source=$%d", i))
		args = append(args, q.Source)
		i++
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	dir := "ASC"
	if q.OrderDesc {
		dir = "DESC"
	}
	query := fmt.Sprintf(`SELECT id, tenant_id, cycle_id, batch_id, project_id, source_id, source,
		payee_party_id, payer_party_id, amount_cents, currency, occurred_at,
		source_fingerprint, metadata, created_at, updated_at
		FROM settlement_entries WHERE %s ORDER BY occurred_at %s, id ASC LIMIT %d OFFSET %d`,
		joinStrings(where, " AND "), dir, pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.SettlementEntry, 0, pageSize)
	for rows.Next() {
		var e domain.SettlementEntry
		var metaJSON []byte
		if err := rows.Scan(&e.ID, &e.TenantID, &e.CycleID, &e.BatchID, &e.ProjectID, &e.SourceID, &e.Source,
			&e.PayeePartyID, &e.PayerPartyID, &e.Amount, &e.Currency, &e.OccurredAt,
			&e.SourceFingerprint, &metaJSON, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		if metaJSON != nil {
			_ = json.Unmarshal(metaJSON, &e.Metadata)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM settlement_entries WHERE %s`, joinStrings(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func (r *EntryRepo) ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.SettlementEntry, error) {
	const q = `SELECT id, tenant_id, cycle_id, batch_id, project_id, source_id, source,
		payee_party_id, payer_party_id, amount_cents, currency, occurred_at,
		source_fingerprint, metadata, created_at, updated_at
		FROM settlement_entries WHERE tenant_id=$1 AND cycle_id=$2 ORDER BY occurred_at ASC, id ASC`
	rows, err := r.pool.Query(ctx, q, tenantID, cycleID)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.SettlementEntry, 0)
	for rows.Next() {
		var e domain.SettlementEntry
		var metaJSON []byte
		if err := rows.Scan(&e.ID, &e.TenantID, &e.CycleID, &e.BatchID, &e.ProjectID, &e.SourceID, &e.Source,
			&e.PayeePartyID, &e.PayerPartyID, &e.Amount, &e.Currency, &e.OccurredAt,
			&e.SourceFingerprint, &metaJSON, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, translateError(err)
		}
		if metaJSON != nil {
			_ = json.Unmarshal(metaJSON, &e.Metadata)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}
	return out, nil
}

// ExceptionRepo is a pgx-backed ExceptionRepo.
type ExceptionRepo struct{ pool *pgxpool.Pool }

func (r *ExceptionRepo) Create(ctx context.Context, e domain.Exception) (domain.Exception, error) {
	if err := e.Validate(); err != nil {
		return domain.Exception{}, err
	}
	snapJSON, err := json.Marshal(e.Snapshot)
	if err != nil {
		return domain.Exception{}, fmt.Errorf("marshal snapshot: %w", err)
	}
	const q = `INSERT INTO exceptions (id, tenant_id, cycle_id, entry_id, rule_version_id, rule_code,
		severity, title, description, hit_reason, status, assignee_id, reporter_id,
		deadline_at, resolved_at, closed_at, snapshot, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		ON CONFLICT (entry_id, rule_code) DO NOTHING`
	_, err = r.pool.Exec(ctx, q, e.ID, e.TenantID, e.CycleID, e.EntryID, e.RuleVersionID, e.RuleCode,
		e.Severity, e.Title, e.Description, e.HitReason, e.Status, e.AssigneeID, e.ReporterID,
		nullableTime(e.DeadlineAt), nullableTime(e.ResolvedAt), nullableTime(e.ClosedAt),
		snapJSON, e.Version, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return domain.Exception{}, translateError(err)
	}
	return e, nil
}

func (r *ExceptionRepo) Get(ctx context.Context, tenantID, id string) (domain.Exception, error) {
	const q = `SELECT id, tenant_id, cycle_id, entry_id, rule_version_id, rule_code,
		severity, title, description, hit_reason, status, assignee_id, reporter_id,
		deadline_at, resolved_at, closed_at, snapshot, version, created_at, updated_at
		FROM exceptions WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var e domain.Exception
	var deadlineAt, resolvedAt, closedAt *time.Time
	var snapJSON []byte
	err := row.Scan(&e.ID, &e.TenantID, &e.CycleID, &e.EntryID, &e.RuleVersionID, &e.RuleCode,
		&e.Severity, &e.Title, &e.Description, &e.HitReason, &e.Status, &e.AssigneeID, &e.ReporterID,
		&deadlineAt, &resolvedAt, &closedAt, &snapJSON, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Exception{}, translateError(err)
	}
	if deadlineAt != nil {
		e.DeadlineAt = *deadlineAt
	}
	if resolvedAt != nil {
		e.ResolvedAt = *resolvedAt
	}
	if closedAt != nil {
		e.ClosedAt = *closedAt
	}
	if snapJSON != nil {
		_ = json.Unmarshal(snapJSON, &e.Snapshot)
	}
	return e, nil
}

func (r *ExceptionRepo) Update(ctx context.Context, e domain.Exception) (domain.Exception, error) {
	if err := e.Validate(); err != nil {
		return domain.Exception{}, err
	}
	snapJSON, err := json.Marshal(e.Snapshot)
	if err != nil {
		return domain.Exception{}, fmt.Errorf("marshal snapshot: %w", err)
	}
	const q = `UPDATE exceptions SET status=$1, assignee_id=$2, deadline_at=$3, resolved_at=$4,
		closed_at=$5, snapshot=$6, version=$7, updated_at=$8
		WHERE tenant_id=$9 AND id=$10 AND version=$11`
	tag, err := r.pool.Exec(ctx, q, e.Status, e.AssigneeID, nullableTime(e.DeadlineAt),
		nullableTime(e.ResolvedAt), nullableTime(e.ClosedAt), snapJSON, e.Version, e.UpdatedAt,
		e.TenantID, e.ID, e.Version-1)
	if err != nil {
		return domain.Exception{}, translateError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Exception{}, domain.NewErrf(domain.CodeAborted, "stale exception %s", e.ID).WithField("version")
	}
	return e, nil
}

func (r *ExceptionRepo) List(ctx context.Context, q application.ExceptionListQuery) ([]domain.Exception, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	if q.CycleID != "" {
		where = append(where, fmt.Sprintf("cycle_id=$%d", i))
		args = append(args, q.CycleID)
		i++
	}
	if q.EntryID != "" {
		where = append(where, fmt.Sprintf("entry_id=$%d", i))
		args = append(args, q.EntryID)
		i++
	}
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status=$%d", i))
		args = append(args, q.Status)
		i++
	}
	if q.Severity != "" {
		where = append(where, fmt.Sprintf("severity=$%d", i))
		args = append(args, q.Severity)
		i++
	}
	if q.AssigneeID != "" {
		where = append(where, fmt.Sprintf("assignee_id=$%d", i))
		args = append(args, q.AssigneeID)
		i++
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, tenant_id, cycle_id, entry_id, rule_version_id, rule_code,
		severity, title, description, hit_reason, status, assignee_id, reporter_id,
		deadline_at, resolved_at, closed_at, snapshot, version, created_at, updated_at
		FROM exceptions WHERE %s ORDER BY
		CASE severity WHEN 'critical' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 WHEN 'low' THEN 1 ELSE 0 END DESC,
		deadline_at ASC NULLS LAST, id ASC LIMIT %d OFFSET %d`,
		joinStrings(where, " AND "), pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.Exception, 0, pageSize)
	for rows.Next() {
		e, err := scanException(rows)
		if err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM exceptions WHERE %s`, joinStrings(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func (r *ExceptionRepo) ListByAssignee(ctx context.Context, tenantID, assigneeID string) ([]domain.Exception, error) {
	const q = `SELECT id, tenant_id, cycle_id, entry_id, rule_version_id, rule_code,
		severity, title, description, hit_reason, status, assignee_id, reporter_id,
		deadline_at, resolved_at, closed_at, snapshot, version, created_at, updated_at
		FROM exceptions WHERE tenant_id=$1 AND assignee_id=$2 ORDER BY id ASC`
	rows, err := r.pool.Query(ctx, q, tenantID, assigneeID)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.Exception, 0)
	for rows.Next() {
		e, err := scanException(rows)
		if err != nil {
			return nil, translateError(err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}
	return out, nil
}

func (r *ExceptionRepo) ListByCycle(ctx context.Context, tenantID, cycleID string) ([]domain.Exception, error) {
	const q = `SELECT id, tenant_id, cycle_id, entry_id, rule_version_id, rule_code,
		severity, title, description, hit_reason, status, assignee_id, reporter_id,
		deadline_at, resolved_at, closed_at, snapshot, version, created_at, updated_at
		FROM exceptions WHERE tenant_id=$1 AND cycle_id=$2 ORDER BY id ASC`
	rows, err := r.pool.Query(ctx, q, tenantID, cycleID)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.Exception, 0)
	for rows.Next() {
		e, err := scanException(rows)
		if err != nil {
			return nil, translateError(err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}
	return out, nil
}

func scanException(rows pgx.Rows) (domain.Exception, error) {
	var e domain.Exception
	var deadlineAt, resolvedAt, closedAt *time.Time
	var snapJSON []byte
	err := rows.Scan(&e.ID, &e.TenantID, &e.CycleID, &e.EntryID, &e.RuleVersionID, &e.RuleCode,
		&e.Severity, &e.Title, &e.Description, &e.HitReason, &e.Status, &e.AssigneeID, &e.ReporterID,
		&deadlineAt, &resolvedAt, &closedAt, &snapJSON, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Exception{}, err
	}
	if deadlineAt != nil {
		e.DeadlineAt = *deadlineAt
	}
	if resolvedAt != nil {
		e.ResolvedAt = *resolvedAt
	}
	if closedAt != nil {
		e.ClosedAt = *closedAt
	}
	if snapJSON != nil {
		_ = json.Unmarshal(snapJSON, &e.Snapshot)
	}
	return e, nil
}

// SummaryRepo is a pgx-backed SummaryRepo.
type SummaryRepo struct{ pool *pgxpool.Pool }

func (r *SummaryRepo) GetLatest(ctx context.Context, tenantID, cycleID string) (domain.Summary, error) {
	const q = `SELECT id, tenant_id, cycle_id, rule_version_id, computed_at, total_entries,
		total_amount_cents, approved_amount_cents, pending_amount_cents,
		exception_count_by_status, exception_count_by_severity, diff_basis, version, created_at
		FROM summaries WHERE tenant_id=$1 AND cycle_id=$2 ORDER BY version DESC LIMIT 1`
	row := r.pool.QueryRow(ctx, q, tenantID, cycleID)
	return scanSummary(row)
}

func (r *SummaryRepo) Save(ctx context.Context, s domain.Summary) (domain.Summary, error) {
	if err := s.Validate(); err != nil {
		return domain.Summary{}, err
	}
	statusJSON, err := json.Marshal(s.ExceptionCountByStatus)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("marshal status counts: %w", err)
	}
	sevJSON, err := json.Marshal(s.ExceptionCountBySeverity)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("marshal severity counts: %w", err)
	}
	diffJSON, err := json.Marshal(s.DiffBasis)
	if err != nil {
		return domain.Summary{}, fmt.Errorf("marshal diff basis: %w", err)
	}
	const q = `INSERT INTO summaries (id, tenant_id, cycle_id, rule_version_id, computed_at,
		total_entries, total_amount_cents, approved_amount_cents, pending_amount_cents,
		exception_count_by_status, exception_count_by_severity, diff_basis, version, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	_, err = r.pool.Exec(ctx, q, s.ID, s.TenantID, s.CycleID, s.RuleVersionID, s.ComputedAt,
		s.TotalEntries, s.TotalAmountCents, s.ApprovedAmountCents, s.PendingAmountCents,
		statusJSON, sevJSON, diffJSON, s.Version, s.ComputedAt)
	if err != nil {
		return domain.Summary{}, translateError(err)
	}
	return s, nil
}

func (r *SummaryRepo) List(ctx context.Context, tenantID, cycleID string, limit int) ([]domain.Summary, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `SELECT id, tenant_id, cycle_id, rule_version_id, computed_at, total_entries,
		total_amount_cents, approved_amount_cents, pending_amount_cents,
		exception_count_by_status, exception_count_by_severity, diff_basis, version, created_at
		FROM summaries WHERE tenant_id=$1 AND cycle_id=$2 ORDER BY version DESC LIMIT $3`
	rows, err := r.pool.Query(ctx, q, tenantID, cycleID, limit)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.Summary, 0)
	for rows.Next() {
		s, err := scanSummary(rows)
		if err != nil {
			return nil, translateError(err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}
	return out, nil
}

func scanSummary(rows pgx.Row) (domain.Summary, error) {
	var s domain.Summary
	var statusJSON, sevJSON, diffJSON []byte
	err := rows.Scan(&s.ID, &s.TenantID, &s.CycleID, &s.RuleVersionID, &s.ComputedAt,
		&s.TotalEntries, &s.TotalAmountCents, &s.ApprovedAmountCents, &s.PendingAmountCents,
		&statusJSON, &sevJSON, &diffJSON, &s.Version, &s.ComputedAt)
	if err != nil {
		return domain.Summary{}, translateError(err)
	}
	if statusJSON != nil {
		_ = json.Unmarshal(statusJSON, &s.ExceptionCountByStatus)
	}
	if sevJSON != nil {
		_ = json.Unmarshal(sevJSON, &s.ExceptionCountBySeverity)
	}
	if diffJSON != nil {
		_ = json.Unmarshal(diffJSON, &s.DiffBasis)
	}
	return s, nil
}

// RecalcRepo is a pgx-backed RecalcRepo.
type RecalcRepo struct{ pool *pgxpool.Pool }

func (r *RecalcRepo) Create(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error) {
	if err := rb.Validate(); err != nil {
		return domain.RecalculationBatch{}, err
	}
	const q = `INSERT INTO recalculation_batches (id, tenant_id, cycle_id, rule_version_id, input_digest,
		trigger_reason, trigger_rule_code, started_at, finished_at, status, output_summary,
		created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`
	var outputJSON []byte
	if rb.OutputSummary.ID != "" {
		b, err := json.Marshal(rb.OutputSummary)
		if err != nil {
			return domain.RecalculationBatch{}, fmt.Errorf("marshal output summary: %w", err)
		}
		outputJSON = b
	}
	_, err := r.pool.Exec(ctx, q, rb.ID, rb.TenantID, rb.CycleID, rb.RuleVersionID, rb.InputDigest,
		rb.TriggerReason, rb.TriggerRuleCode, rb.StartedAt, nullableTime(rb.FinishedAt), rb.Status,
		outputJSON, rb.CreatedAt, rb.UpdatedAt)
	if err != nil {
		return domain.RecalculationBatch{}, translateError(err)
	}
	return rb, nil
}

func (r *RecalcRepo) Update(ctx context.Context, rb domain.RecalculationBatch) (domain.RecalculationBatch, error) {
	if err := rb.Validate(); err != nil {
		return domain.RecalculationBatch{}, err
	}
	var outputJSON []byte
	if rb.OutputSummary.ID != "" {
		b, err := json.Marshal(rb.OutputSummary)
		if err != nil {
			return domain.RecalculationBatch{}, fmt.Errorf("marshal output summary: %w", err)
		}
		outputJSON = b
	}
	const q = `UPDATE recalculation_batches SET input_digest=$1, trigger_reason=$2, trigger_rule_code=$3,
		finished_at=$4, status=$5, output_summary=$6, updated_at=$7 WHERE tenant_id=$8 AND id=$9`
	_, err := r.pool.Exec(ctx, q, rb.InputDigest, rb.TriggerReason, rb.TriggerRuleCode,
		nullableTime(rb.FinishedAt), rb.Status, outputJSON, rb.UpdatedAt, rb.TenantID, rb.ID)
	if err != nil {
		return domain.RecalculationBatch{}, translateError(err)
	}
	return rb, nil
}

func (r *RecalcRepo) Get(ctx context.Context, tenantID, id string) (domain.RecalculationBatch, error) {
	const q = `SELECT id, tenant_id, cycle_id, rule_version_id, input_digest, trigger_reason,
		trigger_rule_code, started_at, finished_at, status, output_summary, created_at, updated_at
		FROM recalculation_batches WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var rb domain.RecalculationBatch
	var finishedAt *time.Time
	var outputJSON []byte
	err := row.Scan(&rb.ID, &rb.TenantID, &rb.CycleID, &rb.RuleVersionID, &rb.InputDigest, &rb.TriggerReason,
		&rb.TriggerRuleCode, &rb.StartedAt, &finishedAt, &rb.Status, &outputJSON, &rb.CreatedAt, &rb.UpdatedAt)
	if err != nil {
		return domain.RecalculationBatch{}, translateError(err)
	}
	if finishedAt != nil {
		rb.FinishedAt = *finishedAt
	}
	if outputJSON != nil {
		_ = json.Unmarshal(outputJSON, &rb.OutputSummary)
	}
	return rb, nil
}

func (r *RecalcRepo) List(ctx context.Context, q application.ListQuery) ([]domain.RecalculationBatch, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	for k, v := range q.Filters {
		col, ok := recalcFilterColumn(k)
		if !ok {
			return nil, application.PageResult{}, domain.NewErrf(domain.CodeInvalidArgument, "unknown filter %s", k).WithField("filter")
		}
		where = append(where, fmt.Sprintf("%s=$%d", col, i))
		args = append(args, v)
		i++
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, tenant_id, cycle_id, rule_version_id, input_digest, trigger_reason,
		trigger_rule_code, started_at, finished_at, status, output_summary, created_at, updated_at
		FROM recalculation_batches WHERE %s ORDER BY started_at DESC, id ASC LIMIT %d OFFSET %d`,
		joinStrings(where, " AND "), pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.RecalculationBatch, 0, pageSize)
	for rows.Next() {
		var rb domain.RecalculationBatch
		var finishedAt *time.Time
		var outputJSON []byte
		if err := rows.Scan(&rb.ID, &rb.TenantID, &rb.CycleID, &rb.RuleVersionID, &rb.InputDigest, &rb.TriggerReason,
			&rb.TriggerRuleCode, &rb.StartedAt, &finishedAt, &rb.Status, &outputJSON, &rb.CreatedAt, &rb.UpdatedAt); err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		if finishedAt != nil {
			rb.FinishedAt = *finishedAt
		}
		if outputJSON != nil {
			_ = json.Unmarshal(outputJSON, &rb.OutputSummary)
		}
		out = append(out, rb)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM recalculation_batches WHERE %s`, joinStrings(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func recalcFilterColumn(k string) (string, bool) {
	switch k {
	case "cycle_id":
		return "cycle_id", true
	case "rule_version_id":
		return "rule_version_id", true
	case "status":
		return "status", true
	}
	return "", false
}

// AnnualRepo is a pgx-backed AnnualRepo.
type AnnualRepo struct{ pool *pgxpool.Pool }

func (r *AnnualRepo) Get(ctx context.Context, tenantID, projectID string, year int) (domain.AnnualAccumulator, error) {
	const q = `SELECT a.project_id, a.year, a.budget_cents, a.disbursed_cents, COALESCE(
		(SELECT json_agg(adj ORDER BY adj.created_at) FROM (
			SELECT id, project_id, year, delta_cents, reason, triggered_by, created_at
			FROM annual_adjustments WHERE project_id=$1 AND year=$2 ORDER BY created_at ASC
		) adj), '[]'::json)
		FROM annual_accumulators a
		WHERE a.project_id=$1 AND a.year=$2`
	row := r.pool.QueryRow(ctx, q, projectID, year)
	var acc domain.AnnualAccumulator
	var adjustmentsJSON []byte
	err := row.Scan(&acc.ProjectID, &acc.Year, &acc.BudgetCents, &acc.DisbursedCents, &adjustmentsJSON)
	if err != nil {
		return domain.AnnualAccumulator{}, translateError(err)
	}
	if adjustmentsJSON != nil {
		_ = json.Unmarshal(adjustmentsJSON, &acc.Adjustments)
	}
	return acc, nil
}

func (r *AnnualRepo) ApplyAdjustment(ctx context.Context, adj domain.Adjustment) (domain.AnnualAccumulator, error) {
	if err := adj.Validate(); err != nil {
		return domain.AnnualAccumulator{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.AnnualAccumulator{}, translateError(err)
	}
	defer tx.Rollback(ctx)
	const upsert = `INSERT INTO annual_accumulators (project_id, year, budget_cents, disbursed_cents)
		VALUES ($1, $2, 0, 0) ON CONFLICT (project_id, year) DO NOTHING`
	if _, err := tx.Exec(ctx, upsert, adj.ProjectID, adj.Year); err != nil {
		return domain.AnnualAccumulator{}, translateError(err)
	}
	const insAdj = `INSERT INTO annual_adjustments (id, project_id, year, delta_cents, reason, triggered_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	if _, err := tx.Exec(ctx, insAdj, adj.ID, adj.ProjectID, adj.Year, adj.DeltaCents, adj.Reason, adj.TriggeredBy, adj.CreatedAt); err != nil {
		return domain.AnnualAccumulator{}, translateError(err)
	}
	const bump = `UPDATE annual_accumulators SET disbursed_cents = disbursed_cents + $1 WHERE project_id=$2 AND year=$3`
	if _, err := tx.Exec(ctx, bump, adj.DeltaCents, adj.ProjectID, adj.Year); err != nil {
		return domain.AnnualAccumulator{}, translateError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.AnnualAccumulator{}, translateError(err)
	}
	return r.Get(ctx, "", adj.ProjectID, adj.Year)
}

func (r *AnnualRepo) ListAdjustments(ctx context.Context, tenantID, projectID string, year int) ([]domain.Adjustment, error) {
	const q = `SELECT id, project_id, year, delta_cents, reason, triggered_by, created_at
		FROM annual_adjustments WHERE project_id=$1 AND year=$2 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, q, projectID, year)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.Adjustment, 0)
	for rows.Next() {
		var a domain.Adjustment
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.Year, &a.DeltaCents, &a.Reason, &a.TriggeredBy, &a.CreatedAt); err != nil {
			return nil, translateError(err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}
	return out, nil
}

// AuditRepo is a pgx-backed AuditRepo.
type AuditRepo struct{ pool *pgxpool.Pool }

func (r *AuditRepo) Append(ctx context.Context, e domain.AuditEntry) (domain.AuditEntry, error) {
	if err := e.Validate(); err != nil {
		return domain.AuditEntry{}, err
	}
	detailJSON, err := json.Marshal(e.Detail)
	if err != nil {
		return domain.AuditEntry{}, fmt.Errorf("marshal detail: %w", err)
	}
	const q = `INSERT INTO audit_entries (id, tenant_id, actor_id, action, entity_type, entity_id, detail, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err = r.pool.Exec(ctx, q, e.ID, e.TenantID, e.ActorID, e.Action, e.EntityType, e.EntityID, detailJSON, e.CreatedAt)
	if err != nil {
		return domain.AuditEntry{}, translateError(err)
	}
	return e, nil
}

func (r *AuditRepo) List(ctx context.Context, q application.ListQuery) ([]domain.AuditEntry, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	for k, v := range q.Filters {
		col, ok := auditFilterColumn(k)
		if !ok {
			return nil, application.PageResult{}, domain.NewErrf(domain.CodeInvalidArgument, "unknown filter %s", k).WithField("filter")
		}
		where = append(where, fmt.Sprintf("%s=$%d", col, i))
		args = append(args, v)
		i++
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, tenant_id, actor_id, action, entity_type, entity_id, detail, created_at
		FROM audit_entries WHERE %s ORDER BY created_at DESC, id ASC LIMIT %d OFFSET %d`,
		joinStrings(where, " AND "), pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.AuditEntry, 0, pageSize)
	for rows.Next() {
		var e domain.AuditEntry
		var detailJSON []byte
		if err := rows.Scan(&e.ID, &e.TenantID, &e.ActorID, &e.Action, &e.EntityType, &e.EntityID, &detailJSON, &e.CreatedAt); err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		if detailJSON != nil {
			_ = json.Unmarshal(detailJSON, &e.Detail)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM audit_entries WHERE %s`, joinStrings(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func auditFilterColumn(k string) (string, bool) {
	switch k {
	case "actor_id":
		return "actor_id", true
	case "action":
		return "action", true
	case "entity_id":
		return "entity_id", true
	case "entity_type":
		return "entity_type", true
	}
	return "", false
}

// UserRepo is a pgx-backed UserRepo.
type UserRepo struct{ pool *pgxpool.Pool }

func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	if err := u.Validate(); err != nil {
		return domain.User{}, err
	}
	const q = `INSERT INTO users (id, tenant_id, username, display_name, email, role, password_hash, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, username) DO NOTHING`
	_, err := r.pool.Exec(ctx, q, u.ID, u.TenantID, u.Username, u.DisplayName, u.Email, u.Role, u.PasswordHash, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return domain.User{}, translateError(err)
	}
	return u, nil
}

func (r *UserRepo) Get(ctx context.Context, tenantID, id string) (domain.User, error) {
	const q = `SELECT id, tenant_id, username, display_name, email, role, password_hash, created_at, updated_at
		FROM users WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var u domain.User
	err := row.Scan(&u.ID, &u.TenantID, &u.Username, &u.DisplayName, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return domain.User{}, translateError(err)
	}
	return u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, tenantID, username string) (domain.User, error) {
	const q = `SELECT id, tenant_id, username, display_name, email, role, password_hash, created_at, updated_at
		FROM users WHERE tenant_id=$1 AND username=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, username)
	var u domain.User
	err := row.Scan(&u.ID, &u.TenantID, &u.Username, &u.DisplayName, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return domain.User{}, translateError(err)
	}
	return u, nil
}

func (r *UserRepo) List(ctx context.Context, q application.ListQuery) ([]domain.User, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	for k, v := range q.Filters {
		col, ok := userFilterColumn(k)
		if !ok {
			return nil, application.PageResult{}, domain.NewErrf(domain.CodeInvalidArgument, "unknown filter %s", k).WithField("filter")
		}
		where = append(where, fmt.Sprintf("%s=$%d", col, i))
		args = append(args, v)
		i++
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, tenant_id, username, display_name, email, role, password_hash, created_at, updated_at
		FROM users WHERE %s ORDER BY username ASC, id ASC LIMIT %d OFFSET %d`,
		joinStrings(where, " AND "), pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.User, 0, pageSize)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Username, &u.DisplayName, &u.Email, &u.Role, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM users WHERE %s`, joinStrings(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func userFilterColumn(k string) (string, bool) {
	switch k {
	case "username":
		return "username", true
	case "role":
		return "role", true
	}
	return "", false
}

func joinStrings(in []string, sep string) string {
	if len(in) == 0 {
		return ""
	}
	out := in[0]
	for _, s := range in[1:] {
		out += sep + s
	}
	return out
}
