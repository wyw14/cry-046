package domain

import (
	"errors"
	"time"
)

type DeliveryStatus string

const (
	DeliveryRequested  DeliveryStatus = "requested"
	DeliveryApproved   DeliveryStatus = "approved"
	DeliveryRejected   DeliveryStatus = "rejected"
	DeliveryExpired    DeliveryStatus = "expired"
	DeliveryDownloaded DeliveryStatus = "downloaded"
)

type DeliveryRequest struct {
	ID           string         `json:"id"`
	PaletteID    string         `json:"palette_id"`
	Applicant    string         `json:"applicant"`
	Format       string         `json:"format"`
	Status       DeliveryStatus `json:"status"`
	ExpiresAt    time.Time      `json:"expires_at"`
	ApprovedBy   string         `json:"approved_by,omitempty"`
	DownloadedAt *time.Time     `json:"downloaded_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	Revision     int            `json:"revision"`
}

func (d DeliveryRequest) CanDownload(now time.Time) error {
	if d.Status != DeliveryApproved && d.Status != DeliveryDownloaded {
		return errors.New("delivery is not approved")
	}
	if d.Status != DeliveryDownloaded && !d.ExpiresAt.After(now) {
		return errors.New("delivery expired")
	}
	return nil
}

func (d *DeliveryRequest) Approve(actor string, now time.Time) error {
	if d.Status != DeliveryRequested {
		return errors.New("delivery is not pending")
	}
	d.Status, d.ApprovedBy, d.Revision = DeliveryApproved, actor, d.Revision+1
	return nil
}

func (d *DeliveryRequest) MarkDownloaded(now time.Time) error {
	if err := d.CanDownload(now); err != nil {
		return err
	}
	d.Status, d.DownloadedAt, d.Revision = DeliveryDownloaded, &now, d.Revision+1
	return nil
}
