package filesystem_asserter

import (
	"sync"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/filesystem_assertion"
	"github.com/codecrafters-io/tester-utils/logger"
)

type FilesystemAsserter struct {
	Timeout    time.Duration
	assertions []filesystem_assertion.FilesystemAssertion
}

func NewFilesystemAsserter(assertions []filesystem_assertion.FilesystemAssertion) *FilesystemAsserter {
	return &FilesystemAsserter{
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
func (a *FilesystemAsserter) RunAssertions(logger *logger.Logger) error {
	if a.Timeout == 0 {
		panic("Codecrafters Internal Error - FilesystemAsserter: Timeout cannot be 0")
	}

	deadline := time.Now().Add(a.Timeout)
	outcomes := runAssertionsConcurrently(a.assertions, deadline)
	logAssertionResultLogs(logger, outcomes)
	return firstAssertionErrorInOrder(outcomes)
}

// runAssertionsConcurrently runs each assertion in its own goroutine and fills outcomes by index.
func runAssertionsConcurrently(assertions []filesystem_assertion.FilesystemAssertion, deadline time.Time) []filesystem_assertion.FilesystemAssertionResult {
	outcomes := make([]filesystem_assertion.FilesystemAssertionResult, len(assertions))
	var waitGroup sync.WaitGroup

	for assertionIndex, assertion := range assertions {
		waitGroup.Add(1)
		go func(idx int, assertion filesystem_assertion.FilesystemAssertion) {
			defer waitGroup.Done()
			outcomes[idx] = runAssertionUntilSuccessOrDeadline(assertion, deadline)
		}(assertionIndex, assertion)
	}

	waitGroup.Wait()
	return outcomes
}

// runAssertionUntilSuccessOrDeadline retries filesystem_assertion.FilesystemAssertion.Run until nil error is returned or deadline is reached.
func runAssertionUntilSuccessOrDeadline(assertion filesystem_assertion.FilesystemAssertion, deadline time.Time) filesystem_assertion.FilesystemAssertionResult {
	var lastResult filesystem_assertion.FilesystemAssertionResult

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

// logAssertionResultLogs logs non-empty success messages in slice order.
func logAssertionResultLogs(logger *logger.Logger, outcomes []filesystem_assertion.FilesystemAssertionResult) {
	for _, outcome := range outcomes {
		// If error is nil, log the success log of that outcome
		if outcome.Err == nil {
			for _, log := range outcome.Logs {
				log.LogMessageUsingLogger(logger)
			}
		}
	}

	for _, outcome := range outcomes {
		if outcome.Err != nil {
			for _, log := range outcome.Logs {
				log.LogMessageUsingLogger(logger)
			}
		}
	}
}

// firstAssertionErrorInOrder returns the first non-nil error in outcomes slice order, or nil if all passed.
func firstAssertionErrorInOrder(outcomes []filesystem_assertion.FilesystemAssertionResult) error {
	for _, outcome := range outcomes {
		if outcome.Err != nil {
			return outcome.Err
		}
	}
	return nil
}
