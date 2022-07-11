package core

import (
	"os"

	"github.com/sirupsen/logrus"
	logrusf "github.com/x-cray/logrus-prefixed-formatter"
)

//------------------------------------------------------------------------------
// Here we define the logging capabilities for this package
//------------------------------------------------------------------------------

// This logger interface does not allow to log errors, since we want errors to created as objects, and be returned to the callers
type Logger interface {
	Info(str string, params ...interface{})
	Debug(str string, params ...interface{})
	Warn(str string, args ...interface{})
}

//------------------------------------------------------------------------------
// Default logger
//------------------------------------------------------------------------------

func (thisComp *ComparisonOptions) SetDefaultLogger() *ComparisonOptions {
	formatter := new(logrusf.TextFormatter)
	formatter.FullTimestamp = true
	formatter.ForceFormatting = true
	formatter.TimestampFormat = "2006-01-02T15:04:05.000000"
	formatter.SetColorScheme(&logrusf.ColorScheme{
		PrefixStyle:    "cyan",
		TimestampStyle: "white+b",
	})

	formatter.ForceColors = true
	formatter.DisableColors = false

	thisComp.logger = &defaultLogger{innerLogger: &logrus.Logger{
		Level:     logrus.DebugLevel,
		Formatter: formatter,
		Out:       os.Stderr,
	}}

	return thisComp
}

type defaultLogger struct {
	innerLogger *logrus.Logger
}

func (thisLogger *defaultLogger) Info(str string, params ...interface{}) {
	thisLogger.innerLogger.Infof(str, params...)
}

func (thisLogger *defaultLogger) Debug(str string, params ...interface{}) {
	thisLogger.innerLogger.Debugf(str, params...)
}

func (thisLogger *defaultLogger) Warn(str string, params ...interface{}) {
	thisLogger.innerLogger.Warnf(str, params...)
}
