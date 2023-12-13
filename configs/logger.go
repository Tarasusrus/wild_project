package configs

import (
	"context"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

// todo ну это пизда, надо переделывать
type GormLoggerAdapter struct {
	*log.Logger
	logLevel logger.LogLevel
}

func NewGormLoggerAdapter() logger.Interface {
	return &GormLoggerAdapter{
		Logger:   log.New(os.Stdout, "\r\n", log.LstdFlags),
		logLevel: logger.Info, // Установите начальный уровень логирования
	}
}

func (l *GormLoggerAdapter) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

// Реализация других методов с учетом уровня логирования
func (l *GormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.Printf("INFO: "+msg, data...)
	}
}

func (l *GormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.Printf("WARN: "+msg, data...)
	}
}

func (l *GormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.Printf("ERROR: "+msg, data...)
	}
}

func (l *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel >= logger.Info {
		elapsed := time.Since(begin)
		sql, rows := fc()
		l.Printf("TRACE: %s [%vms] - Rows affected: %d - Error: %v", sql, float64(elapsed.Milliseconds()), rows, err)
	}
}
