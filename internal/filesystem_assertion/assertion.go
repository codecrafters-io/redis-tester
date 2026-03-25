package filesystem_assertion

import (
	"fmt"
	"slices"

	"github.com/codecrafters-io/tester-utils/logger"
)

const (
	_DEBUG   = "DEBUG"
	_INFO    = "INFO"
	_SUCCESS = "SUCCESS"
	_ERROR   = "ERROR"
)

type FilesystemAssertionLog struct {
	Type    string
	Message string
}

func (l *FilesystemAssertionLog) LogMessageUsingLogger(logger *logger.Logger) {
	switch l.Type {
	case _DEBUG:
		logger.Debugf("%s", l.Message)
	case _INFO:
		logger.Infof("%s", l.Message)
	case _SUCCESS:
		logger.Successf("%s", l.Message)
	case _ERROR:
		logger.Errorf("%s", l.Message)
	default:
		panic(fmt.Sprintf("Codecrafters Internal Error - %s is not a valid log type", l.Type))
	}
}

func NewFilesystemAssertionLog(logType string, message string) FilesystemAssertionLog {
	if !slices.Contains([]string{_DEBUG, _INFO, _SUCCESS, _ERROR}, logType) {
		panic(fmt.Sprintf("Codecrafters Internal Error - %s is not a valid log type", logType))
	}

	return FilesystemAssertionLog{
		Type:    logType,
		Message: message,
	}
}

type FilesystemAssertionResult struct {
	Logs []FilesystemAssertionLog
	Err  error
}

type FilesystemAssertion interface {
	Run() FilesystemAssertionResult
}
