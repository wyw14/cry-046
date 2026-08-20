package domain

import (
	"testing"
	"time"
)

func TestDeliveryExpiry(t *testing.T) {
	d := DeliveryRequest{Status: DeliveryApproved, ExpiresAt: time.Unix(100, 0)}
	if err := d.CanDownload(time.Unix(101, 0)); err == nil {
		t.Fatal("expired delivery allowed")
	}
}
