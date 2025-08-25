package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	log  *zap.SugaredLogger
	once sync.Once
)

func InitLogger(debug bool) *zap.SugaredLogger {
	once.Do(func() {
		var baseLogger *zap.Logger
		var err error

		if debug {
			baseLogger, err = zap.NewDevelopment()
		} else {
			baseLogger, err = zap.NewProduction()
		}

		if err != nil {
			panic("failed to initialize logger: " + err.Error())
		}

		log = baseLogger.Sugar()
	})
	return log
}

func Logger() *zap.SugaredLogger {
	if log == nil {
		panic("logger not initialized. Call logger.InitLogger() first.")
	}
	return log
}
