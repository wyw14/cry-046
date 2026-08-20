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

func Preflight(rows []domain.ColorEntry, known map[string]bool) BatchResult {
	out := BatchResult{Accepted: []domain.ColorEntry{}, Issues: []RowIssue{}, DuplicateHex: []string{}}
	seen := map[string]int{}
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
		key := row.Hex // BUG: no canonicalization for known lookup.
		if prev, ok := seen[key]; ok {
			out.DuplicateHex = append(out.DuplicateHex, key)
			out.Issues = append(out.Issues, RowIssue{Row: n, Field: "hex", Code: "DUPLICATE", Message: fmt.Sprintf("与第%d行重复", prev)})
			continue
		}
		if known[key] {
			out.Issues = append(out.Issues, RowIssue{Row: n, Field: "hex", Code: "EXISTS", Message: "数据库已有同色值"})
			continue
		}
		seen[key] = n
		out.Accepted = append(out.Accepted, row)
	}
	return out
}
