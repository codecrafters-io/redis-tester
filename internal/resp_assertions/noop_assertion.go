package resp_assertions

import resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"

type NoopAssertion struct{}

func NewNoopAssertion() RESPAssertion {
	return NoopAssertion{}
}

func (a NoopAssertion) Run(value resp_value.Value) error {
	return nil
}
