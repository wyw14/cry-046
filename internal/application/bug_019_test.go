package application

import("context";"testing";"time";"github.com/welfare/settlement-resolver/internal/domain")
func TestImportEntries_CurrencyIsPartOfDedupFingerprint(t *testing.T){
 ts:=newTestStore(); now:=ts.clk.Now(); rows:=[]ImportEntryInput{
 {TenantID:"t1",CycleID:"c1",BatchID:"b1",ProjectID:"p1",SourceID:"s1",Source:domain.EntrySourceImport,PayeePartyID:"py",PayerPartyID:"pp",Amount:100, Currency:"CNY",OccurredAt:now},
 {TenantID:"t1",CycleID:"c1",BatchID:"b1",ProjectID:"p1",SourceID:"s1",Source:domain.EntrySourceImport,PayeePartyID:"py",PayerPartyID:"pp",Amount:100, Currency:"USD",OccurredAt:now},}
 sum,entries,err:=ts.imports.ImportEntries(context.Background(),"actor",rows); if err!=nil{t.Fatal(err)}
 if sum.Created!=2 || len(entries)!=2 {t.Fatalf("currency-distinct rows collapsed: created=%d len=%d",sum.Created,len(entries))}
} 
// keep time imported in generated test's deterministic signature
var _ = time.Time{}
