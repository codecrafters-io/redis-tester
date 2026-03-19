package filesystem_assertion

import (
	"sync"
	"time"

	"github.com/codecrafters-io/tester-utils/logger"
)

type FileSystemAsserter struct {
	Timeout    time.Duration
	assertions []FilesystemAssertion
}

func NewFileSystemAsserter(assertions []FilesystemAssertion) *FileSystemAsserter {
	return &FileSystemAsserter{
		// Default timeout for FS asserter
		Timeout:    2 * time.Second,
		assertions: assertions,
	}
}

// RunAssertions runs all assertions concurrently.
// Each assertion is run until either it returns no error, or timeout expires
// After either all assertions have returned no error, or timeout expires,
// The accumulated success logs (of assertions which passed) are logged
// If there are any errors, the first error from the a.assertions slice is returned to preserve order
func (a *FileSystemAsserter) RunAssertions(logger *logger.Logger) error {
	if a.Timeout == 0 {
		panic("Codecrafters Internal Error - FilesystemAsserter: Timeout cannot be 0")
	}

	deadline := time.Now().Add(a.Timeout)
	outcomes := runAssertionsConcurrently(a.assertions, deadline)
	logAssertionSuccessMessagesInOrder(logger, outcomes)
	return firstAssertionErrorInOrder(outcomes)
}

// runAssertionsConcurrently runs each assertion in its own goroutine and fills outcomes by index.
func runAssertionsConcurrently(assertions []FilesystemAssertion, deadline time.Time) []FileSystemAssertionResult {
	outcomes := make([]FileSystemAssertionResult, len(assertions))
	var waitGroup sync.WaitGroup

	for assertionIndex, filesystemAssertion := range assertions {
		waitGroup.Add(1)
		go func(idx int, assertion FilesystemAssertion) {
			defer waitGroup.Done()
			outcomes[idx] = runAssertionUntilSuccessOrDeadline(assertion, deadline)
		}(assertionIndex, filesystemAssertion)
	}

	waitGroup.Wait()
	return outcomes
}

// runAssertionUntilSuccessOrDeadline retries filesystemAssertion.Run until nil error is returned or deadline is reached.
func runAssertionUntilSuccessOrDeadline(assertion FilesystemAssertion, deadline time.Time) FileSystemAssertionResult {
	var lastResult FileSystemAssertionResult

	for {
		result := assertion.Run()
		lastResult = result

		if result.Err == nil {
			return result
		}

		if time.Now().After(deadline) {
			return lastResult
		}

		// Sleep 10ms instead of 1ms because fs operations can take longer
		time.Sleep(10 * time.Millisecond)
	}
}

// logAssertionSuccessMessagesInOrder logs non-empty success messages in slice order.
func logAssertionSuccessMessagesInOrder(logger *logger.Logger, outcomes []FileSystemAssertionResult) {
	for _, outcome := range outcomes {
		// If error is nil, log the success log of that outcome
		if outcome.Err == nil {
			logger.Successf("%s", outcome.SuccessLog)
		}
	}
}

// firstAssertionErrorInOrder returns the first non-nil error in outcomes slice order, or nil if all passed.
func firstAssertionErrorInOrder(outcomes []FileSystemAssertionResult) error {
	for _, outcome := range outcomes {
		if outcome.Err != nil {
			return outcome.Err
		}
	}
	return nil
}
