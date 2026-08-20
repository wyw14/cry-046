package domain
import "testing"
func TestDeriveDoesNotAliasEntries(t *testing.T){p:=Palette{ID:"p",Name:"v",Status:StatusDelivered,Version:2,Entries:[]ColorEntry{{Name:"red",Hex:"#112233"}}};d,e:=p.Derive("d","derived");if e!=nil{t.Fatal(e)};d.Entries[0].Hex="#ffffff";if p.Entries[0].Hex==d.Entries[0].Hex{t.Fatal("derived palette aliases source entries")}}
