package resp_assertions

import (
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type RESPAssertion interface {
	Run(value resp_value.Value) error
}
