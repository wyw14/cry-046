package validation

import (
	"fmt"
	"github.com/wyw14/cry-046/internal/domain"
	"regexp"
	"strings"
)

type RowIssue struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
type BatchResult struct {
	Accepted     []domain.ColorEntry `json:"accepted"`
	Issues       []RowIssue          `json:"issues"`
	DuplicateHex []string            `json:"duplicate_hex"`
}

var nameRule = regexp.MustCompile(`^[\p{Han}A-Za-z0-9 _-]{1,80}$`)

// canonicalHex returns the lowercased form of a hex color so that values
// differing only by letter case (e.g. #FF0000 and #ff0000) are recognized as
// the same color during duplicate and existing-color checks.
func canonicalHex(hex string) string {
	return strings.ToLower(strings.TrimSpace(hex))
}

// normalizeKnown lowercases every key in the known-color map so that
// existing-color lookups are case-insensitive regardless of how the caller
// stored the hex values.
func normalizeKnown(known map[string]bool) map[string]bool {
	if known == nil {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(known))
	for k, v := range known {
		if v {
			out[canonicalHex(k)] = true
		}
	}
	return out
}

func Preflight(rows []domain.ColorEntry, known map[string]bool) BatchResult {
	out := BatchResult{Accepted: []domain.ColorEntry{}, Issues: []RowIssue{}, DuplicateHex: []string{}}
	seen := map[string]int{}
	knownColors := normalizeKnown(known)
	for i, row := range rows {
		n := i + 1
		if !nameRule.MatchString(row.Name) {
			out.Issues = append(out.Issues, RowIssue{Row: n, Field: "name", Code: "INVALID_NAME", Message: "名称包含未允许字符"})
			continue
		}
		if !strings.HasPrefix(row.Hex, "#") || len(row.Hex) != 7 {
			out.Issues = append(out.Issues, RowIssue{Row: n, Field: "hex", Code: "INVALID_HEX", Message: "色值必须为六位十六进制"})
			continue
		}
		// Compare on the canonical (lowercased) hex so case-only differences
		// do not bypass the duplicate or existing-color checks.
		key := canonicalHex(row.Hex)
		if prev, ok := seen[key]; ok {
			out.DuplicateHex = append(out.DuplicateHex, key)
			out.Issues = append(out.Issues, RowIssue{Row: n, Field: "hex", Code: "DUPLICATE", Message: fmt.Sprintf("与第%d行重复", prev)})
			continue
		}
		if knownColors[key] {
			out.Issues = append(out.Issues, RowIssue{Row: n, Field: "hex", Code: "EXISTS", Message: "数据库已有同色值"})
			continue
		}
		seen[key] = n
		out.Accepted = append(out.Accepted, row)
	}
	return out
}
