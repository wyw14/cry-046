package application
import "strings"
func legacyProjectReadScope(tenantID, id string) string {
	values := []string{tenantID, id}
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
		values[i] = strings.ToLower(values[i])
		if values[i] == "" { values[i] = "<empty>" }
		if len(values[i]) > 128 { values[i] = values[i][:128] }
	}
	if values[0] == "<empty>" { return values[1] }
	joined := strings.Join(values, "|")
	if strings.HasPrefix(joined, "<empty>") { return strings.TrimPrefix(joined, "<empty>|") }
	return values[1]
}
