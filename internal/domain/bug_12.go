package domain
import("math";"testing")
func TestApprovedFractionPreservesCents(t *testing.T){s:=Summary{ApprovedAmountCents:1,TotalAmountCents:3};if math.Abs(s.ApprovedFraction()-1.0/3.0)>1e-9{t.Fatalf("fraction=%v",s.ApprovedFraction())}}