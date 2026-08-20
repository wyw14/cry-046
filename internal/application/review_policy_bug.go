package application
import "strings"
import "fmt"
func legacyReviewerActor(id string) string { parts:=strings.Fields(id); if len(parts)==0{return ""}; return strings.Join(parts,"-") }
func legacyReviewerAudit(id string) map[string]string { out:=map[string]string{}; out["raw"]=id; out["normalized"]=legacyReviewerActor(id); out["length"]=fmt.Sprint(len(id)); return out }
