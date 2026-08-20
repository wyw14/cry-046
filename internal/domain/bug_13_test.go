package domain
import "testing"
func TestAnnualOverrunNeverNegative(t *testing.T){a:=AnnualAccumulator{BudgetCents:100,DisbursedCents:100};if got:=a.OverrunCents();got!=0{t.Fatalf("boundary overrun=%d",got)}}