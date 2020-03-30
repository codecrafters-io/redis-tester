package internal

import (
	"log"
	"os"

	"github.com/fatih/color"
)

func colorize(colorToUse color.Attribute, fstring string, args ...interface{}) string {
	return color.New(colorToUse).SprintfFunc()(fstring, args...)
}

func debugColorize(fstring string, args ...interface{}) string {
	return colorize(color.FgCyan, fstring, args...)
}

func infoColorize(fstring string, args ...interface{}) string {
	return colorize(color.FgBlue, fstring, args...)
}

func successColorize(fstring string, args ...interface{}) string {
	return colorize(color.FgGreen, fstring, args...)
}

func errorColorize(fstring string, args ...interface{}) string {
	return colorize(color.FgRed, fstring, args...)
}

func yellowColorize(fstring string, args ...interface{}) string {
	return colorize(color.FgYellow, fstring, args...)
}

type customLogger struct {
	logger  log.Logger
	isDebug bool
	isQuiet bool // Only CRITICAL logs
}

func getLogger(isDebug bool, prefix string) *customLogger {
	color.NoColor = false

	prefix = yellowColorize(prefix)
	return &customLogger{
		logger:  *log.New(os.Stdout, prefix, 0),
		isDebug: isDebug,
	}
}

func getQuietLogger(prefix string) *customLogger {
	color.NoColor = false

	prefix = yellowColorize(prefix)
	return &customLogger{
		logger:  *log.New(os.Stdout, prefix, 0),
		isDebug: false,
		isQuiet: true,
	}
}

func (l *customLogger) Successf(fstring string, args ...interface{}) {
	if l.isQuiet {
		return
	}
	msg := successColorize(fstring, args...)
	l.Successln(msg)
}

func (l *customLogger) Successln(msg string) {
	if l.isQuiet {
		return
	}
	msg = successColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Infof(fstring string, args ...interface{}) {
	if l.isQuiet {
		return
	}
	msg := infoColorize(fstring, args...)
	l.Infoln(msg)
}

func (l *customLogger) Infoln(msg string) {
	if l.isQuiet {
		return
	}
	msg = infoColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Criticalf(fstring string, args ...interface{}) {
	if !l.isQuiet {
		panic("Critical is only for quiet loggers")
	}
	msg := errorColorize(fstring, args...)
	l.Criticalln(msg)
}

func (l *customLogger) Criticalln(msg string) {
	if !l.isQuiet {
		panic("Critical is only for quiet loggers")
	}
	msg = errorColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Errorf(fstring string, args ...interface{}) {
	if l.isQuiet {
		return
	}
	msg := errorColorize(fstring, args...)
	l.Errorln(msg)
}

func (l *customLogger) Errorln(msg string) {
	if l.isQuiet {
		return
	}
	msg = errorColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Debugf(fstring string, args ...interface{}) {
	if !l.isDebug {
		return
	}
	msg := debugColorize(fstring, args...)
	l.Debugln(msg)
}

func (l *customLogger) Debugln(msg string) {
	if !l.isDebug {
		return
	}
	msg = debugColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Plainln(msg string) {
	l.logger.Println(msg)
}
