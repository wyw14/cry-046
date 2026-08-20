package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr, StorageRoot, PostgresDSN string
	MaxUploadBytes                     int64
}

func Load() Config {
	max, _ := strconv.ParseInt(env("MAX_UPLOAD_BYTES", "5242880"), 10, 64)
	return Config{HTTPAddr: env("HTTP_ADDR", ":8090"), StorageRoot: env("STORAGE_ROOT", "./var/deliveries"), PostgresDSN: env("POSTGRES_DSN", "postgres://palette:palette@localhost:5432/palette?sslmode=disable"), MaxUploadBytes: max}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
