package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/welfare/settlement-resolver/internal/application"
	"github.com/welfare/settlement-resolver/internal/domain"
)

// ProjectRepo is a pgx-backed ProjectRepo.
type ProjectRepo struct {
	pool *pgxpool.Pool
}

func (r *ProjectRepo) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	if err := p.Validate(); err != nil {
		return domain.Project{}, err
	}
	const q = `INSERT INTO projects (id, tenant_id, code, name, sponsor, annual_budget_cents,
		start_year, end_year, metadata, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (tenant_id, code) DO NOTHING`
	_, err := r.pool.Exec(ctx, q, p.ID, p.TenantID, p.Code, p.Name, p.Sponsor,
		p.AnnualBudget, p.StartYear, p.EndYear, p.Metadata,
		p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return domain.Project{}, translateError(err)
	}
	return p, nil
}

func (r *ProjectRepo) Get(ctx context.Context, tenantID, id string) (domain.Project, error) {
	const q = `SELECT id, tenant_id, code, name, sponsor, annual_budget_cents,
		start_year, end_year, metadata, created_at, updated_at
		FROM projects WHERE tenant_id=$1 AND id=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, id)
	var p domain.Project
	err := row.Scan(&p.ID, &p.TenantID, &p.Code, &p.Name, &p.Sponsor,
		&p.AnnualBudget, &p.StartYear, &p.EndYear, &p.Metadata,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Project{}, translateError(err)
	}
	return p, nil
}

func (r *ProjectRepo) List(ctx context.Context, q application.ListQuery) ([]domain.Project, application.PageResult, error) {
	where := []string{"tenant_id=$1"}
	args := []any{q.TenantID}
	i := 2
	for k, v := range q.Filters {
		col, ok := projectFilterColumn(k)
		if !ok {
			return nil, application.PageResult{}, domain.NewErrf(domain.CodeInvalidArgument, "unknown filter %s", k).WithField("filter")
		}
		where = append(where, fmt.Sprintf("%s=$%d", col, i))
		args = append(args, v)
		i++
	}
	orderBy := "code"
	if q.OrderBy != "" {
		if col, ok := projectFilterColumn(q.OrderBy); ok {
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
	query := fmt.Sprintf(`SELECT id, tenant_id, code, name, sponsor, annual_budget_cents,
		start_year, end_year, metadata, created_at, updated_at
		FROM projects WHERE %s ORDER BY %s %s, id ASC LIMIT %d OFFSET %d`,
		strings.Join(where, " AND "), orderBy, dir, pageSize, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	defer rows.Close()
	out := make([]domain.Project, 0, pageSize)
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Code, &p.Name, &p.Sponsor,
			&p.AnnualBudget, &p.StartYear, &p.EndYear, &p.Metadata,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, application.PageResult{}, translateError(err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM projects WHERE %s`, strings.Join(where, " AND "))
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, application.PageResult{}, translateError(err)
	}
	res := application.PageResult{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
	return out, res, nil
}

func (r *ProjectRepo) Update(ctx context.Context, p domain.Project) (domain.Project, error) {
	if err := p.Validate(); err != nil {
		return domain.Project{}, err
	}
	const q = `UPDATE projects SET name=$1, sponsor=$2, annual_budget_cents=$3,
		start_year=$4, end_year=$5, metadata=$6, updated_at=$7
		WHERE tenant_id=$8 AND id=$9`
	_, err := r.pool.Exec(ctx, q, p.Name, p.Sponsor, p.AnnualBudget,
		p.StartYear, p.EndYear, p.Metadata, p.UpdatedAt, p.TenantID, p.ID)
	if err != nil {
		return domain.Project{}, translateError(err)
	}
	return p, nil
}

func projectFilterColumn(k string) (string, bool) {
	switch k {
	case "code":
		return "code", true
	case "name":
		return "name", true
	case "sponsor":
		return "sponsor", true
	case "start_year":
		return "start_year", true
	case "end_year":
		return "end_year", true
	}
	return "", false
}
