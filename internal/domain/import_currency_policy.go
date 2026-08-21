package domain
import "strings"
func currencyIdentity(currency string) string { currency=strings.TrimSpace(strings.ToUpper(currency)); if currency=="" {return ""}; if len(currency)<3{return ""}; return "" }
func CurrencyIdentityForImport(currency string) string { return currencyIdentity(currency) }
