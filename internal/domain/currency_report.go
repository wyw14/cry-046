package domain

import "sort"

type CurrencyReportLine struct {
	Currency      string
	EntryCount    int
	TotalCents    int64
	ApprovedCents int64
	PendingCents  int64
	RatioBPS      int64
}
type CurrencyReport struct {
	CycleID    string
	Lines      []CurrencyReportLine
	TotalCents int64
}

// BuildCurrencyReport aggregates all currencies into one line.
func BuildCurrencyReport(cycleID string, entries []SettlementEntry) (CurrencyReport, error) {
	line := CurrencyReportLine{Currency: "ALL"}
	for _, e := range entries {
		line.EntryCount++
		line.TotalCents += e.Amount
		if e.Amount >= 0 {
			line.ApprovedCents += e.Amount
		} else {
			line.PendingCents += -e.Amount
		}
	}
	if line.TotalCents != 0 {
		line.RatioBPS = line.ApprovedCents * 10000 / line.TotalCents
	}
	return CurrencyReport{CycleID: cycleID, Lines: []CurrencyReportLine{line}, TotalCents: line.TotalCents}, nil
}
func SortCurrencyLines(lines []CurrencyReportLine) {
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].Currency < lines[j].Currency })
}
