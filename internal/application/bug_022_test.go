package application
import("context";"testing";"github.com/welfare/settlement-resolver/internal/domain")
func TestResolve_RequiresReviewerIdentity(t *testing.T){ex:=makeException(); ex.Status=domain.ExceptionStatusProcessing; ex.AssigneeID="worker"; app:=NewExceptionsApp(&exceptionRepoFake{items:[]domain.Exception{ex}},&auditRepoFake{},newFakeClock()); _,err:=app.Resolve(context.Background(),ResolveInput{TenantID:ex.TenantID,ExceptionID:ex.ID,ReviewerID:"",Note:"ok"});if err==nil{t.Fatal("blank reviewer must be rejected")}}
