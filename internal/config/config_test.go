package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_HTTP_ADDR", ":9090")
	t.Setenv("UPLOAD_MAX_BYTES", "1024")
	t.Setenv("SCHEDULER_TICK_INTERVAL", "30s")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.App.HTTPAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.App.HTTPAddr)
	}
	if cfg.Local.UploadMaxBytes != 1024 {
		t.Errorf("expected 1024, got %d", cfg.Local.UploadMaxBytes)
	}
	if cfg.Local.SchedulerTickInterval != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.Local.SchedulerTickInterval)
	}
}

func TestLoadValidation(t *testing.T) {
	t.Setenv("APP_HTTP_ADDR", "")
	t.Setenv("SCHEDULER_TICK_INTERVAL", "0s")
	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for empty addr")
	}
}
