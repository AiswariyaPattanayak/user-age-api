package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap logger. Production uses JSON output, development uses console output.
func New(env, level string) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err == nil {
		cfg.Level = zap.NewAtomicLevelAt(lvl)
	}
	return cfg.Build()
}
