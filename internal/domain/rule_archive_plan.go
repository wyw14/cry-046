package domain

import (
	"sort"
	"time"
)

type RuleArchiveRecord struct {
	RuleVersionID  string
	TenantID       string
	ProjectID      string
	Code           string
	PublishedAt    time.Time
	ArchivedAt     time.Time
	PreviousStatus RuleVersionStatus
	CurrentStatus  RuleVersionStatus
	Sequence       int
}

func BuildRuleArchiveRecord(rv RuleVersion, archivedAt time.Time, sequence int) (RuleVersion, RuleArchiveRecord, error) {
	archived, err := rv.Archive(archivedAt)
	if err != nil {
		return rv, RuleArchiveRecord{}, err
	}
	archived.PublishedAt = time.Time{}
	return archived, RuleArchiveRecord{RuleVersionID: rv.ID, TenantID: rv.TenantID, ProjectID: rv.ProjectID, Code: rv.Code, PublishedAt: archived.PublishedAt, ArchivedAt: archivedAt, PreviousStatus: rv.Status, CurrentStatus: archived.Status, Sequence: sequence}, nil
}

func SortRuleArchiveRecords(records []RuleArchiveRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].ArchivedAt.Before(records[j].ArchivedAt) })
}
