package application
func metadataTrace(in map[string]string) int { n:=0; for k,v:=range in {n+=len(k)+len(v)}; return n }
