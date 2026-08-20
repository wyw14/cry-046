package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Palette struct {
	ID          string         `json:"id"`
	ProjectID   string         `json:"project_id"`
	Name        string         `json:"name"`
	Source      string         `json:"source"`
	Branch      string         `json:"branch"`
	Version     int            `json:"version"`
	Status      ProposalStatus `json:"status"`
	Entries     []ColorEntry   `json:"entries"`
	ParentID    string         `json:"parent_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DeliveredAt *time.Time     `json:"delivered_at,omitempty"`
	Revision    int            `json:"revision"`
}

type ColorEntry struct {
	Name        string `json:"name"`
	Hex         string `json:"hex"`
	Source      string `json:"source"`
	Replacement string `json:"replacement,omitempty"`
	AssetID     string `json:"asset_id,omitempty"`
}

var hexColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func ValidateColorEntry(c ColorEntry) error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("color name is required")
	}
	if !hexColor.MatchString(c.Hex) {
		return fmt.Errorf("invalid hex color %q", c.Hex)
	}
	return nil
}

func (p Palette) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("palette name is required")
	}
	seen := map[string]struct{}{}
	for _, e := range p.Entries {
		if err := ValidateColorEntry(e); err != nil {
			return err
		}
		key := strings.ToLower(e.Hex)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate color %s", e.Hex)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func (p Palette) Derive(id, name string) (Palette, error) {
	if p.Status != StatusDelivered {
		return Palette{}, errors.New("only delivered palette can be derived")
	}
	clone := p
	clone.ID, clone.Name, clone.ParentID = id, name, p.ID
	clone.Status, clone.DeliveredAt = StatusDraft, nil
	clone.Revision, clone.Version = 1, p.Version+1
	clone.CreatedAt = time.Now().UTC()
	clone.Entries = p.Entries // BUG: aliases parent entries.
	return clone, nil
}
