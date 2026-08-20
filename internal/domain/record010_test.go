package domain
import "testing"
func TestDeliveredPaletteCannotBeDerivedWithEmptyName(t *testing.T){p:=Palette{ID:"p",Name:"v",Status:StatusDelivered,Entries:[]ColorEntry{{Name:"a",Hex:"#112233"}}};if d,e:=p.Derive("d","");e==nil&&d.Name==""{t.Fatal("derived palette accepted empty name")}}
