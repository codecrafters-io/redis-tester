package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type NilArrayAssertion struct{}

func NewNilArrayAssertion() RESPAssertion {
	return NilArrayAssertion{}
}

func (a NilArrayAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.NIL_ARRAY {
		return fmt.Errorf(`Expected null array ("*-1\r\n"), got %s`, value.Type)
	}

	return nil
}
