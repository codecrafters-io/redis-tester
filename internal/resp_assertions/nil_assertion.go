package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type NilAssertion struct{}

func NewNilAssertion() RESPAssertion {
	return NilAssertion{}
}

func (a NilAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.NIL {
		return fmt.Errorf(`Expected null bulk string ("$-1\r\n"), got %s`, value.Type)
	}

	return nil
}
