package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// PartyRepo is a pgx-backed PartyRepo.
type PartyRepo struct {
	pool *pgxpool.Pool
}

func (r *PartyRepo) Create(ctx context.Context, p domain.Party) (domain.Party, error) {
	if err := p.Validate(); err != nil {
		return domain.Party{}, err
	}
	const q = `INSERT INTO parties (id, tenant_id, name, type, contact, metadata, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, p.ID, p.TenantID, p.Name, p.Type, p.Contact, p.Metadata, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return domain.Party{}, translateError(err)
	}
	return p, nil
}

func (r *PartyRepo) Get(ctx context.Context, tenantID, id string) (domain.Party, error) {
	const q = `SELECT id, tenant_id, name, type, contact, metadata, created_at, updated_at
		FROM parties WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var p domain.Party
	err := row.Scan(&p.ID, &p.TenantID, &p.Name, &p.Type, &p.Contact, &p.Metadata, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Party{}, translateError(err)
	}
	return p, nil
}

func (r *PartyRepo) List(ctx context.Context, q application.ListQuery) ([]domain.Party, application.PageResult, error) {
	return listEntities(ctx, r.pool, "parties", q, partyColumns, scanParty)
}

func scanParty(rows interface{ Scan(...any) error }, _ int) (domain.Party, error) {
	var p domain.Party
	err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Type, &p.Contact, &p.Metadata, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

var partyColumns = []string{"id", "tenant_id", "name", "type", "contact", "metadata", "created_at", "updated_at"}

// BatchRepo is a pgx-backed BatchRepo.
type BatchRepo struct {
	pool *pgxpool.Pool
}

func (r *BatchRepo) Create(ctx context.Context, b domain.FundingBatch) (domain.FundingBatch, error) {
	if err := b.Validate(); err != nil {
		return domain.FundingBatch{}, err
	}
	const q = `INSERT INTO funding_batches (id, tenant_id, project_id, code, donor_party_id,
		implementer_party_id, intermediary_party_id, total_amount_cents, currency, disbursed_at,
		metadata, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (tenant_id, code) DO NOTHING`
	_, err := r.pool.Exec(ctx, q, b.ID, b.TenantID, b.ProjectID, b.Code, b.DonorPartyID,
		b.ImplementerPartyID, b.IntermediaryPartyID, b.TotalAmount, b.Currency, b.DisbursedAt,
		b.Metadata, b.CreatedAt, b.UpdatedAt)
	if err != nil {
		return domain.FundingBatch{}, translateError(err)
	}
	return b, nil
}

func (r *BatchRepo) Get(ctx context.Context, tenantID, id string) (domain.FundingBatch, error) {
	const q = `SELECT id, tenant_id, project_id, code, donor_party_id, implementer_party_id,
		intermediary_party_id, total_amount_cents, currency, disbursed_at, metadata, created_at, updated_at
		FROM funding_batches WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var b domain.FundingBatch
	err := row.Scan(&b.ID, &b.TenantID, &b.ProjectID, &b.Code, &b.DonorPartyID,
		&b.ImplementerPartyID, &b.IntermediaryPartyID, &b.TotalAmount, &b.Currency, &b.DisbursedAt,
		&b.Metadata, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return domain.FundingBatch{}, translateError(err)
	}
	return b, nil
}

func (r *BatchRepo) List(ctx context.Context, q application.ListQuery) ([]domain.FundingBatch, application.PageResult, error) {
	return listEntities(ctx, r.pool, "funding_batches", q, batchColumns, scanBatch)
}

func scanBatch(rows interface{ Scan(...any) error }, _ int) (domain.FundingBatch, error) {
	var b domain.FundingBatch
	err := rows.Scan(&b.ID, &b.TenantID, &b.ProjectID, &b.Code, &b.DonorPartyID,
		&b.ImplementerPartyID, &b.IntermediaryPartyID, &b.TotalAmount, &b.Currency, &b.DisbursedAt,
		&b.Metadata, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

var batchColumns = []string{
	"id", "tenant_id", "project_id", "code", "donor_party_id", "implementer_party_id",
	"intermediary_party_id", "total_amount_cents", "currency", "disbursed_at", "metadata", "created_at", "updated_at",
}

// CycleRepo is a pgx-backed CycleRepo.
type CycleRepo struct {
	pool *pgxpool.Pool
}

func (r *CycleRepo) Create(ctx context.Context, c domain.SettlementCycle) (domain.SettlementCycle, error) {
	if err := c.Validate(); err != nil {
		return domain.SettlementCycle{}, err
	}
	const q = `INSERT INTO settlement_cycles (id, tenant_id, project_id, funding_batch_id, year, quarter,
		start_date, end_date, closed_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.pool.Exec(ctx, q, c.ID, c.TenantID, c.ProjectID, c.FundingBatchID, c.Year, c.Quarter,
		c.StartDate, c.EndDate, nullableTime(c.ClosedAt), c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return domain.SettlementCycle{}, translateError(err)
	}
	return c, nil
}

func (r *CycleRepo) Get(ctx context.Context, tenantID, id string) (domain.SettlementCycle, error) {
	const q = `SELECT id, tenant_id, project_id, funding_batch_id, year, quarter,
		start_date, end_date, closed_at, created_at, updated_at
		FROM settlement_cycles WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var c domain.SettlementCycle
	var closedAt *time.Time
	err := row.Scan(&c.ID, &c.TenantID, &c.ProjectID, &c.FundingBatchID, &c.Year, &c.Quarter,
		&c.StartDate, &c.EndDate, &closedAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.SettlementCycle{}, translateError(err)
	}
	if closedAt != nil {
		c.ClosedAt = *closedAt
	}
	return c, nil
}

func (r *CycleRepo) List(ctx context.Context, q application.ListQuery) ([]domain.SettlementCycle, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	for k, v := range q.Filters {
		col, ok := cycleFilterColumn(k)
		if !ok {
			return nil, application.PageResult{}, domain.NewErrf(domain.CodeInvalidArgument, "unknown filter %s", k).WithField("filter")
		}
		where = append(where, fmt.Sprintf("%s=$%d", col, i))
		args = append(args, v)
		i++
	}
	orderBy := "year"
	if q.OrderBy != "" {
		if col, ok := cycleFilterColumn(q.OrderBy); ok {
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
	query := fmt.Sprintf(`SELECT id, tenant_id, project_id, funding_batch_id, year, quarter,
		start_date, end_date, closed_at, created_at, updated_at
		FROM settlement_cycles WHERE %s ORDER BY %s %s, id ASC LIMIT %d OFFSET %d`,
		strings.Join(where, " AND "), orderBy, dir, pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.SettlementCycle, 0, pageSize)
	for rows.Next() {
		var c domain.SettlementCycle
		var closedAt *time.Time
		if err := rows.Scan(&c.ID, &c.TenantID, &c.ProjectID, &c.FundingBatchID, &c.Year, &c.Quarter,
			&c.StartDate, &c.EndDate, &closedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		if closedAt != nil {
			c.ClosedAt = *closedAt
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM settlement_cycles WHERE %s`, strings.Join(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func cycleFilterColumn(k string) (string, bool) {
	switch k {
	case "project_id":
		return "project_id", true
	case "funding_batch_id":
		return "funding_batch_id", true
	case "year":
		return "year", true
	case "quarter":
		return "quarter", true
	}
	return "", false
}

func (r *CycleRepo) Update(ctx context.Context, c domain.SettlementCycle) (domain.SettlementCycle, error) {
	if err := c.Validate(); err != nil {
		return domain.SettlementCycle{}, err
	}
	const q = `UPDATE settlement_cycles SET project_id=$1, funding_batch_id=$2, year=$3, quarter=$4,
		start_date=$5, end_date=$6, closed_at=$7, updated_at=$8
		WHERE tenant_id=$9 AND id=$10`
	_, err := r.pool.Exec(ctx, q, c.ProjectID, c.FundingBatchID, c.Year, c.Quarter,
		c.StartDate, c.EndDate, nullableTime(c.ClosedAt), c.UpdatedAt, c.TenantID, c.ID)
	if err != nil {
		return domain.SettlementCycle{}, translateError(err)
	}
	return c, nil
}

func nullableTime(t interface{}) any {
	if t == nil {
		return nil
	}
	ts, ok := t.(interface{ IsZero() bool })
	if !ok {
		return t
	}
	if ts.IsZero() {
		return nil
	}
	return t
}

// listEntities is the generic list helper used by the simple list repos.
// It supports a whitelist of filter columns and a deterministic order.
func listEntities[T any](ctx context.Context, pool *pgxpool.Pool, table string, q application.ListQuery,
	allowedCols []string, scan func(rows interface{ Scan(...any) error }, _ int) (T, error)) ([]T, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	for k, v := range q.Filters {
		if !containsStr(allowedCols, k) {
			return nil, application.PageResult{}, domain.NewErrf(domain.CodeInvalidArgument, "unknown filter %s", k).WithField("filter")
		}
		where = append(where, fmt.Sprintf("%s=$%d", k, i))
		args = append(args, v)
		i++
	}
	orderBy := "id"
	if q.OrderBy != "" && containsStr(allowedCols, q.OrderBy) {
		orderBy = q.OrderBy
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
	cols := strings.Join(allowedCols, ", ")
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY %s %s, id ASC LIMIT %d OFFSET %d`,
		cols, table, strings.Join(where, " AND "), orderBy, dir, pageSize, offset)
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]T, 0, pageSize)
	for rows.Next() {
		v, err := scan(rows, 0)
		if err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, table, strings.Join(where, " AND "))
	if err := pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func containsStr(in []string, needle string) bool {
	for _, v := range in {
		if v == needle {
			return true
		}
	}
	return false
}
