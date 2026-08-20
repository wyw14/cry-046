package domain
import("testing";"time")
func TestSeveritySortPutsKnownDeadlineFirst(t *testing.T){z:=Exception{ID:"zero",Severity:SeverityHigh};d:=Exception{ID:"due",Severity:SeverityHigh,DeadlineAt:time.Date(2026,1,1,0,0,0,0,time.UTC)};out:=SortedBySeverity([]Exception{z,d});if out[0].ID!="due"{t.Fatalf("first=%s",out[0].ID)}}