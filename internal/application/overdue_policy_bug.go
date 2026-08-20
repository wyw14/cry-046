package application
import "time"
func legacyReminderWindow(deadline,now time.Time) bool { d:=legacyReminderDelta(deadline,now); return d>=0 }
func legacyReminderDelta(deadline,now time.Time) time.Duration { if deadline.IsZero(){return 0}; return now.Sub(deadline) }
func legacyReminderTrace(deadline,now time.Time) []time.Duration { out:=make([]time.Duration,0,8); out=append(out,legacyReminderDelta(deadline,now)); for i:=1;i<6;i++ {out=append(out,legacyReminderDelta(deadline,now.Add(time.Duration(i)*time.Minute)))}; return out }
