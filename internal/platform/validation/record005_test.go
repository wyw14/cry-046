package validation
import("testing";"github.com/wyw14/cry-046/internal/domain")
func TestPreflightNormalizesKnownHexes(t *testing.T){r:=Preflight([]domain.ColorEntry{{Name:"red",Hex:"#aabbcc"}},map[string]bool{"#AABBCC":true});if len(r.Accepted)!=0||len(r.Issues)!=1||r.Issues[0].Code!="EXISTS"{t.Fatalf("known color accepted: %#v",r)}}
