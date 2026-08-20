package domain

import (
	"testing"
	"time"
)

func TestAssetExportable(t *testing.T) {
	now := time.Unix(1000, 0)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	cases := []struct {
		name string
		asset Asset
		want bool
	}{
		{"active with complete authorization and future expiry", Asset{State: AssetActive, CopyrightNote: "licensed", ExpiresAt: &future}, true},
		{"active with complete authorization and no expiry", Asset{State: AssetActive, CopyrightNote: "licensed", ExpiresAt: nil}, true},
		{"active but expired", Asset{State: AssetActive, CopyrightNote: "licensed", ExpiresAt: &past}, false},
		{"active but expiry is exactly now", Asset{State: AssetActive, CopyrightNote: "licensed", ExpiresAt: &now}, false},
		{"active but missing copyright note", Asset{State: AssetActive, CopyrightNote: "   ", ExpiresAt: &future}, false},
		{"revoked despite valid authorization", Asset{State: AssetRevoked, CopyrightNote: "licensed", ExpiresAt: &future}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.asset.Exportable(now); got != c.want {
				t.Fatalf("Exportable=%v, want %v", got, c.want)
			}
		})
	}
}
