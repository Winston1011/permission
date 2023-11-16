package logger

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

// LogLevel
type LogLevel int

const (
	Debug LogLevel = iota + 1
	Info
	Error
	Warn
)

// Interface logger interface
type Interface interface {
	Logger(*gin.Context, LogLevel, string)
}

// default Logger
var Default = &DefaultLogger{
	writer: log.New(os.Stdout, "\r\n", log.LstdFlags),
}

type DefaultLogger struct {
	writer *log.Logger
}

func (d *DefaultLogger) Logger(ctx *gin.Context, level LogLevel, msg string) {
	switch level {
	case Debug:
		fallthrough
	case Info:
		fallthrough
	case Warn:
		fallthrough
	case Error:
		fallthrough
	default:
		d.writer.Print(msg)
	}
}
