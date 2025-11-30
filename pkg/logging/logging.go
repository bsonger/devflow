package logging

import (
	"github.com/bsonger/devflow/pkg/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init() {
	var cfg zap.Config

	if model.C.Log.Format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	level := zapcore.InfoLevel
	_ = level.Set(model.C.Log.Level)
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	Logger = logger
}
