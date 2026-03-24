package filesystem_assertion

import (
	"fmt"
	"slices"

	"github.com/codecrafters-io/tester-utils/logger"
)

// FileSystemAssertionResult is the outcome of a single filesystem assertion run.
// A result, when obtained, is used by the asserter in the following manner:
// Print InfoLogs in order
// Print SuccessLog if Err is nil, Err if not nil

const (
	_DEBUG   = "DEBUG"
	_INFO    = "INFO"
	_SUCCESS = "SUCCESS"
)

type FileSystemAssertionLog struct {
	Type    string
	Message string
}

func (l *FileSystemAssertionLog) LogMessageUsingLogger(logger *logger.Logger) {
	switch l.Type {
	case _DEBUG:
		logger.Debugf("%s", l.Message)
	case _INFO:
		logger.Infof("%s", l.Message)
	case _SUCCESS:
		logger.Successf("%s", l.Message)
	default:
		panic(fmt.Sprintf("Codecrafters Internal Error - %s is not a valid log type", l.Type))
	}
}

func NewFileSystemAssertionResultLog(logType string, message string) FileSystemAssertionLog {
	if !slices.Contains([]string{_DEBUG, _INFO, _SUCCESS}, logType) {
		panic(fmt.Sprintf("Codecrafters Internal Error - %s is not a valid log type", logType))
	}

	return FileSystemAssertionLog{
		Type:    logType,
		Message: message,
	}
}

type FileSystemAssertionResult struct {
	Logs []FileSystemAssertionLog
	Err  error
}

type FilesystemAssertion interface {
	Run() FileSystemAssertionResult
}
