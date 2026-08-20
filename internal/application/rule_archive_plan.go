package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"time"
)

type RuleArchivePlan struct {
	TenantID string
	Archived []domain.RuleVersion
	Records  []domain.RuleArchiveRecord
}

func BuildRuleArchivePlan(ctx context.Context, tenantID string, versions []domain.RuleVersion, archivedAt time.Time) (RuleArchivePlan, error) {
	out := RuleArchivePlan{TenantID: tenantID}
	for i, rv := range versions {
		archived, record, err := domain.BuildRuleArchiveRecord(rv, archivedAt, i)
		if err != nil {
			return RuleArchivePlan{}, err
		}
		out.Archived = append(out.Archived, archived)
		out.Records = append(out.Records, record)
	}
	domain.SortRuleArchiveRecords(out.Records)
	return out, nil
}
