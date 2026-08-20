// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the resolved runtime configuration for the platform.
type Config struct {
	App    AppConfig
	DB     DBConfig
	CORS   CORSConfig
	Local  LocalAdapterConfig
	Seed   SeedConfig
	Logger LoggerConfig
}

type AppConfig struct {
	Name            string
	Env             string
	HTTPAddr        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type LocalAdapterConfig struct {
	StorageDir              string
	UploadMaxBytes          int64
	UploadAllowedTypes      []string
	SchedulerTickInterval   time.Duration
	OverdueReminderInterval time.Duration
}

type SeedConfig struct {
	DemoData     bool
	DemoProjects int
	DemoBatches  int
}

type LoggerConfig struct {
	Level    string
	Encoding string
}

// Load reads the configuration from the process environment.
func Load() (Config, error) {
	cfg := Config{
		App: AppConfig{
			Name:            getenv("APP_NAME", "welfare-settlement-resolver"),
			Env:             getenv("APP_ENV", "development"),
			HTTPAddr:        getenv("APP_HTTP_ADDR", ":8080"),
			ReadTimeout:     getenvDuration("APP_HTTP_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getenvDuration("APP_HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getenvDuration("APP_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getenvDuration("APP_SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		DB: DBConfig{
			DSN:             getenv("DB_DSN", "postgres://resolver:resolver@127.0.0.1:5432/welfare_resolver?sslmode=disable"),
			MaxOpenConns:    getenvInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getenvInt("DB_MAX_IDLE_CONNS", 2),
			ConnMaxLifetime: getenvDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins: getenvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		Local: LocalAdapterConfig{
			StorageDir:              getenv("LOCAL_STORAGE_DIR", "./storage"),
			UploadMaxBytes:          int64(getenvInt("UPLOAD_MAX_BYTES", 5*1024*1024)),
			UploadAllowedTypes:      getenvList("UPLOAD_ALLOWED_TYPES", []string{"image/png", "image/jpeg", "application/pdf", "text/csv", "text/plain"}),
			SchedulerTickInterval:   getenvDuration("SCHEDULER_TICK_INTERVAL", 15*time.Second),
			OverdueReminderInterval: getenvDuration("OVERDUE_REMINDER_INTERVAL", 10*time.Minute),
		},
		Seed: SeedConfig{
			DemoData:     getenvBool("SEED_DEMO_DATA", true),
			DemoProjects: getenvInt("SEED_DEMO_PROJECTS", 3),
			DemoBatches:  getenvInt("SEED_DEMO_BATCHES", 2),
		},
		Logger: LoggerConfig{
			Level:    getenv("LOG_LEVEL", "info"),
			Encoding: getenv("LOG_ENCODING", "json"),
		},
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	if c.App.HTTPAddr == "" {
		return fmt.Errorf("APP_HTTP_ADDR must not be empty")
	}
	if c.Local.UploadMaxBytes <= 0 {
		return fmt.Errorf("UPLOAD_MAX_BYTES must be positive")
	}
	if c.Local.SchedulerTickInterval <= 0 {
		return fmt.Errorf("SCHEDULER_TICK_INTERVAL must be positive")
	}
	if c.Local.OverdueReminderInterval <= 0 {
		return fmt.Errorf("OVERDUE_REMINDER_INTERVAL must be positive")
	}
	if c.Seed.DemoProjects < 0 || c.Seed.DemoBatches < 0 {
		return fmt.Errorf("demo seed counts must be non-negative")
	}
	return nil
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getenvList(key string, def []string) []string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return def
}
