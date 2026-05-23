package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	global = l
}

func Logger() *zap.Logger { return global }

func Sync() { _ = global.Sync() }
