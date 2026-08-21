package application

import (
	"context"
	"github.com/welfare/settlement-resolver/internal/domain"
	"testing"
)

func TestCurrencyReportSeparatesCurrenciesAndPreservesTotals(t *testing.T) {
	entries := []domain.SettlementEntry{{ID: "cny1", Currency: "CNY", Amount: 100}, {ID: "usd1", Currency: "USD", Amount: 200}, {ID: "cny2", Currency: "CNY", Amount: -20}}
	got, err := BuildCurrencySettlementReport(context.Background(), "c1", entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Lines) != 2 {
		t.Fatalf("currencies merged: %+v", got)
	}
	if got.Lines[0].Currency != "CNY" || got.Lines[0].TotalCents != 80 || got.Lines[1].Currency != "USD" || got.Lines[1].TotalCents != 200 {
		t.Fatalf("currency lines wrong: %+v", got.Lines)
	}
}
