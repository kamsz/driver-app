package common

import (
	"strings"

	nsq "github.com/bitly/go-nsq"
	log "github.com/sirupsen/logrus"
)

var (
	nsqDebugLevel = nsq.LogLevelDebug.String()
	nsqInfoLevel  = nsq.LogLevelInfo.String()
	nsqWarnLevel  = nsq.LogLevelWarning.String()
	nsqErrLevel   = nsq.LogLevelError.String()
)

// NSQLogrusLogger is an adaptor between the weird go-nsq Logger and our
// standard logrus logger.
type NSQLogrusLogger struct{}

// NewNSQLogrusLogger returns a new NSQLogrusLogger and the current log level.
// This is a format to easily plug into nsq.SetLogger.
func NewNSQLogrusLogger() (logger NSQLogrusLogger, level nsq.LogLevel) {
	return NewNSQLogrusLoggerAtLevel(log.GetLevel())
}

// NewNSQLogrusLoggerAtLevel returns a new NSQLogrusLogger with the provided log level mapped to nsq.LogLevel for easily plugging into nsq.SetLogger.
func NewNSQLogrusLoggerAtLevel(l log.Level) (logger NSQLogrusLogger, level nsq.LogLevel) {
	logger = NSQLogrusLogger{}
	level = nsq.LogLevelWarning
	switch l {
	case log.DebugLevel:
		level = nsq.LogLevelDebug
	case log.InfoLevel:
		level = nsq.LogLevelInfo
	case log.WarnLevel:
		level = nsq.LogLevelWarning
	case log.ErrorLevel:
		level = nsq.LogLevelError
	}
	return
}

// Output implements stdlib log.Logger.Output using logrus
// Decodes the go-nsq log messages to figure out the log level
func (n NSQLogrusLogger) Output(_ int, s string) error {
	if len(s) > 3 {
		msg := strings.TrimSpace(s[3:])
		switch s[:3] {
		case nsqDebugLevel:
			log.Debugln(msg)
		case nsqInfoLevel:
			log.Infoln(msg)
		case nsqWarnLevel:
			log.Warnln(msg)
		case nsqErrLevel:
			log.Errorln(msg)
		default:
			log.Infoln(msg)
		}
	}
	return nil
}
