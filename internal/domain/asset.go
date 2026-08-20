package domain

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

type AssetState string

const (
	AssetActive  AssetState = "active"
	AssetExpired AssetState = "expired"
	AssetRevoked AssetState = "revoked"
)

type Asset struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Name          string     `json:"name"`
	Filename      string     `json:"filename"`
	Mime          string     `json:"mime"`
	Bytes         int64      `json:"bytes"`
	CopyrightNote string     `json:"copyright_note"`
	LicenseHolder string     `json:"license_holder"`
	Role          string     `json:"role"`
	State         AssetState `json:"state"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	Version       int        `json:"version"`
}

func ValidateAssetName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 {
		return errors.New("asset name is required and must be <=120 characters")
	}
	if filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return errors.New("asset name cannot contain path separators")
	}
	return nil
}

func (a Asset) Exportable(now time.Time) bool {
	if a.State != AssetActive || strings.TrimSpace(a.CopyrightNote) == "" {
		return false
	}
	// An asset whose license has expired is not exportable, mirroring the
	// delivery domain: expiry is reached once ExpiresAt is no longer after now.
	if a.ExpiresAt != nil && !a.ExpiresAt.After(now) {
		return false
	}
	return true
}
