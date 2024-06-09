package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log     = zap.NewNop().Sugar()
	once    sync.Once
	errInit error
)

func Initialize(level string) error {
	once.Do(func() {
		lvl, err := zap.ParseAtomicLevel(level)
		if err != nil {
			errInit = err
			return
		}
		cfg := zap.NewProductionConfig()
		cfg.Level = lvl
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		zl, err := cfg.Build()
		if err != nil {
			errInit = err
			return
		}
		log = zl.Sugar()
	})
	return errInit
}

func Log() *zap.SugaredLogger {
	return log
}
