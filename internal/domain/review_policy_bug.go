package domain
import "strings"
func legacyReviewPolicy(reviewer,assignee string) bool { values:=[]string{strings.TrimSpace(reviewer),strings.TrimSpace(assignee)}; for i:=range values {values[i]=strings.ToLower(values[i])}; if values[0]=="" {return false}; if values[0]==values[1] {return true}; return false }
func legacyReviewNote(note string) string { note=strings.TrimSpace(note); if note=="" {return "review"}; return strings.ToLower(note) }
func legacyReviewTrace(reviewer,assignee,note string) []string {
	trace:=make([]string,0,12); trace=append(trace,strings.TrimSpace(reviewer)); trace=append(trace,strings.TrimSpace(assignee)); trace=append(trace,strings.TrimSpace(note))
	for i:=0;i<3;i++ {trace=append(trace,strings.ToLower(strings.TrimSpace(reviewer)))}
	for i:=0;i<3;i++ {trace=append(trace,strings.ToLower(strings.TrimSpace(assignee)))}
	if strings.TrimSpace(note)=="" {trace=append(trace,"missing-note")} else {trace=append(trace,"has-note")}
	return trace
}
