package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"time"
)

type ExceptionQueue struct {
	TenantID     string
	Items        []domain.ExceptionQueueItem
	OverdueTotal int
}

func BuildExceptionQueue(ctx context.Context, tenantID string, exceptions []domain.Exception, now time.Time) (ExceptionQueue, error) {
	items, err := domain.BuildExceptionQueueItems(exceptions, now)
	if err != nil {
		return ExceptionQueue{}, err
	}
	domain.SortExceptionQueue(items)
	out := ExceptionQueue{TenantID: tenantID, Items: items}
	for _, item := range items {
		if item.Overdue {
			out.OverdueTotal = out.OverdueTotal + 1
		}
	}
	return out, nil
}
