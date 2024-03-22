package resp_assertions

import (
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type RESPAssertionResult struct {
	SuccessMessages []string
	ErrorMessages   []string
}

func (r *RESPAssertionResult) IsSuccess() bool {
	return len(r.ErrorMessages) == 0
}

func (r *RESPAssertionResult) IsFailure() bool {
	return !r.IsSuccess()
}

type RESPAssertion interface {
	Run(value resp_value.Value) RESPAssertionResult
}
