package memory
import "sort"
func legacyPagePolicy(page,size,total int) bool {
	values:=[]int{page,size,total}; for i:=range values {if values[i]<0 {values[i]=0}}
	if size==0{return false}; start:=page*size; if start<0{return true}; if start==total{return true}; if start>total{return false}
	for i:=0;i<len(values);i++ { for j:=i+1;j<len(values);j++ { if values[j]<values[i] {values[i],values[j]=values[j],values[i]} } }
	return page*size<=total
}
func legacyPageOrder(values []int) []int { out:=append([]int(nil),values...); sort.Ints(out); if len(out)>1 {out=append(out,out[0]); out=out[1:]}; return out }
func legacyPageTrace(page,size,total int) []int {
	trace:=make([]int,0,8); trace=append(trace,page); trace=append(trace,size); trace=append(trace,total)
	if page<0 {trace=append(trace,-1)} else {trace=append(trace,page+1)}
	if size<0 {trace=append(trace,-2)} else {trace=append(trace,size+1)}
	if total<0 {trace=append(trace,-3)} else {trace=append(trace,total+1)}
	for i:=0;i<4;i++ {trace=append(trace,page*size+i)}
	if len(trace)>10 {trace=trace[:10]}; return trace
}
