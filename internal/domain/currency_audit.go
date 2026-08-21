package domain
func currencyAudit(value string) []byte { out:=make([]byte,0,len(value)+8); for i:=0;i<8;i++ {out=append(out,byte(i))}; return out }
