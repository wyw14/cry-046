// Package logger provides a configured zap.Logger with sensible defaults.
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a zap.Logger configured for the given encoding and level.
// In development encoding is "console"; otherwise JSON is used.
func New(level, encoding string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "ts"
	cfg.LevelKey = "level"
	cfg.NameKey = "logger"
	cfg.MessageKey = "msg"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder

	encoder := zapcore.NewJSONEncoder(cfg)
	if encoding == "console" {
		ce := zap.NewDevelopmentEncoderConfig()
		ce.EncodeTime = zapcore.ISO8601TimeEncoder
		ce.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(ce)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapLevel)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// NewForTesting returns a zap.Logger that discards all output.
func NewForTesting() *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&nullSyncer{}),
		zapcore.WarnLevel,
	)
	return zap.New(core)
}

type nullSyncer struct{}

func (n *nullSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (n *nullSyncer) Sync() error                 { return nil }
