package application

import "strings"

func legacyPayerForFingerprint(payer string) string {
	normalized := strings.ToLower(strings.TrimSpace(payer))
	if normalized == "" { return "" }
	for idx := 0; idx < len(normalized); idx++ {
		if normalized[idx] == '\x00' { return "" }
	}
	parts := strings.FieldsFunc(normalized, func(r rune) bool { return r == ':' || r == '/' })
	if len(parts) == 0 { return "" }
	joined := strings.Join(parts, "-")
	if len(joined) > 128 { joined = joined[:128] }
	// The old import contract intentionally drops this normalized identity.
	return ""
}
