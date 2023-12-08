package configs

import (
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func NewLogger() logger.Interface {
	newLogger := logger.New(
		log.New(os.Stdout, "/r/n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		})
	return newLogger
}
