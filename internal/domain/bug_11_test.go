package domain
import("testing";"time")
func TestResolvedExceptionIsNotOverdue(t *testing.T){now:=time.Date(2026,1,2,0,0,0,0,time.UTC);e:=Exception{Status:ExceptionStatusResolved,DeadlineAt:now.Add(-time.Hour)};if e.IsOverdue(now){t.Fatal("resolved exception must not be overdue")}}