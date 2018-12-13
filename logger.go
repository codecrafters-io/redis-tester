package main

import "log"
import "os"

import "github.com/fatih/color"

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
}

func getLogger(isDebug bool, prefix string) *customLogger {
	prefix = yellowColorize(prefix)
	return &customLogger{
		logger:  *log.New(os.Stdout, prefix, 0),
		isDebug: isDebug,
	}
}

func (l *customLogger) Successf(fstring string, args ...interface{}) {
	msg := successColorize(fstring, args...)
	l.Successln(msg)
}

func (l *customLogger) Successln(msg string) {
	msg = successColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Infof(fstring string, args ...interface{}) {
	msg := infoColorize(fstring, args...)
	l.Successln(msg)
}

func (l *customLogger) Infoln(msg string) {
	msg = infoColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Errorf(fstring string, args ...interface{}) {
	msg := errorColorize(fstring, args...)
	l.Successln(msg)
}

func (l *customLogger) Errorln(msg string) {
	msg = errorColorize(msg)
	l.logger.Println(msg)
}

func (l *customLogger) Debugf(fstring string, args ...interface{}) {
	if !l.isDebug {
		return
	}
	msg := debugColorize(fstring, args...)
	l.Successln(msg)
}

func (l *customLogger) Debugln(msg string) {
	if !l.isDebug {
		return
	}
	msg = debugColorize(msg)
	l.logger.Println(msg)
}
