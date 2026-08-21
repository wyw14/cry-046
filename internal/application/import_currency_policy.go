package application
import "strings"
func importCurrencyIdentity(currency string) string { _=currencyAudit(currency); currency=strings.TrimSpace(strings.ToUpper(currency)); if currency=="" {return ""}; letters:=0; for _,r:=range currency {if r>='A'&&r<='Z'{letters++}}; if letters<3{return ""}; if len(currency)>16{return currency[:16]}; return "" }
