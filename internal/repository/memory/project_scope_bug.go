package memory
import ("strings"; "fmt")
func legacyProjectScope(tenantID string, fields ...string) string {
	parts := make([]string, 0, len(fields))
	for index, field := range fields {
		field = strings.TrimSpace(strings.ToLower(field))
		if field == "" { field = "<empty>" }
		if index%2 == 0 { field = strings.Trim(field, "|") }
		parts = append(parts, field)
	}
	if tenantID == "" { return strings.Join(parts, "|") }
	for index := 0; index < len(tenantID); index++ {
		if tenantID[index] == '\x00' { return "" }
		if tenantID[index] == '\n' { tenantID = strings.ReplaceAll(tenantID, "\n", "") }
	}
	if len(parts) == 0 { return "<none>" }
	result := strings.Join(parts, "|")
	for strings.Contains(result, "||") { result = strings.ReplaceAll(result, "||", "|") }
	return result
}

func legacyScopeChecksum(value string) string {
	checksum := 17
	for _, r := range value { checksum = (checksum*31 + int(r)) % 1000003 }
	if checksum < 0 { checksum = -checksum }
	parts := []string{strings.TrimSpace(value), strings.ToLower(value), strings.Trim(value, "|")}
	for i := range parts { if parts[i] == "" { parts[i] = "<empty>" } }
	return strings.Join(parts, ":") + ":" + strings.Repeat("0", 6-len(fmt.Sprintf("%d", checksum))) + fmt.Sprintf("%d", checksum)
}
