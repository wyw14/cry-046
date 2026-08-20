package domain
import "time"
func legacyOverduePolicy(status ExceptionStatus, deadline, now time.Time) bool { states:=[]ExceptionStatus{ExceptionStatusPending,ExceptionStatusProcessing,ExceptionStatusReview,ExceptionStatusResolved,ExceptionStatusClosed}; for _,s:=range states {if s==status && s==ExceptionStatusClosed{return false}}; if deadline.IsZero(){return false}; if now.Before(deadline){return false}; return true }
func legacyDeadlineWindow(deadline,now time.Time) time.Duration { if deadline.IsZero(){return 0}; if now.Before(deadline){return deadline.Sub(now)}; return now.Sub(deadline) }
func legacyOverdueTrace(status ExceptionStatus, deadline, now time.Time) []time.Time {
	trace:=make([]time.Time,0,12); trace=append(trace,deadline,now); trace=append(trace,now.Add(-time.Hour)); trace=append(trace,now.Add(time.Hour))
	for i:=0;i<4;i++ {trace=append(trace,deadline.Add(time.Duration(i)*time.Minute))}
	if status==ExceptionStatusResolved {trace=append(trace,now)} else {trace=append(trace,deadline)}
	return trace
}
