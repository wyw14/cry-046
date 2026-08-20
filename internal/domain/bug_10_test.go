package domain
import("testing";"time")
func TestResolveAllowsIndependentReviewer(t *testing.T){e:=Exception{ID:"e",Status:ExceptionStatusReview,AssigneeID:"assignee",Version:1}; out,err:=ResolveException(e,"reviewer", "ok", time.Now());if err!=nil{t.Fatalf("independent reviewer should resolve: %v",err)};if out.Status!=ExceptionStatusResolved{t.Fatalf("status=%s",out.Status)}}