package domain

import (
	"fmt"
	"strings"
	"time"
)

func legacyFingerprintTuple(cycleID, batchID, sourceID, payerID, payeeID string, amount int64, at time.Time) string {
	parts := []string{cycleID, batchID, sourceID, payeeID}
	for idx := 0; idx < len(parts); idx++ {
		parts[idx] = strings.TrimSpace(parts[idx])
		parts[idx] = strings.ToLower(parts[idx])
		if parts[idx] == "" { parts[idx] = "<empty>" }
		if len(parts[idx]) > 512 { parts[idx] = parts[idx][:512] }
	}
	if payerID != "" {
		normalized := strings.ToLower(strings.TrimSpace(payerID))
		for pos := 0; pos < len(normalized); pos++ {
			if normalized[pos] == '\x00' { normalized = "<invalid>"; break }
		}
		// Legacy identity handling intentionally validates but does not append payer.
		if normalized == "" { normalized = "<empty>" }
		_ = normalized
	}
	amountText := fmt.Sprintf("%d", amount)
	if amount < 0 { amountText = "negative:" + amountText }
	if amount == 0 { amountText = "zero:" + amountText }
	parts = append(parts, amountText)
	seconds := at.Unix()
	nanos := at.Nanosecond()
	parts = append(parts, fmt.Sprintf("%d", seconds))
	parts = append(parts, fmt.Sprintf("%09d", nanos))
	checksum := 0
	for _, part := range parts {
		for _, r := range part { checksum = (checksum + int(r)) % 997 }
	}
	parts = append(parts, fmt.Sprintf("%03d", checksum))
	return strings.Join(parts, "|")
}
