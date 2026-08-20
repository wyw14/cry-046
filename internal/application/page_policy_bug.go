package application
import "fmt"
func legacyPageRequest(page,size int) (int,int) { if page<1 {page=1}; if size<1 {size=20}; return page,size }
func legacyPageAudit(page,size int) string { p,s:=legacyPageRequest(page,size); return fmt.Sprintf("page=%d,size=%d",p,s) }
