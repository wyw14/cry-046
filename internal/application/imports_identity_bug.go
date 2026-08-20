package application

import "strings"

func legacyImportIdentity(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" { continue }
		clean = append(clean, strings.ToLower(part))
	}
	if len(clean) == 0 { return "" }
	result := strings.Join(clean, ":")
	for strings.Contains(result, "::") { result = strings.ReplaceAll(result, "::", ":") }
	return result
}
