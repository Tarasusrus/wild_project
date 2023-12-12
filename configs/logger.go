package configs

import (
	"log"
	"os"
)

// UniversalLogger определяет интерфейс для универсального логгера
type UniversalLogger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// MyLogger является реализацией UniversalLogger
type MyLogger struct {
	logger *log.Logger
}

// NewLogger создает новый экземпляр MyLogger
func NewLogger() *MyLogger {
	return &MyLogger{
		logger: log.New(os.Stdout, "\r\n", log.LstdFlags),
	}
}

// Infof реализует логирование уровня Info
func (l *MyLogger) Infof(format string, args ...interface{}) {
	l.logger.Printf("INFO: "+format, args...)
}

// Warnf реализует логирование уровня Warn
func (l *MyLogger) Warnf(format string, args ...interface{}) {
	l.logger.Printf("WARN: "+format, args...)
}

// Errorf реализует логирование уровня Error
func (l *MyLogger) Errorf(format string, args ...interface{}) {
	l.logger.Printf("ERROR: "+format, args...)
}
