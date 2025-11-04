package resp_assertions

import (
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type NilArrayAssertion struct{}

func NewNilArrayAssertion() RESPAssertion {
	return NilArrayAssertion{}
}

func (a NilArrayAssertion) Run(value resp_value.Value) error {
	nilArrayTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.NIL_ARRAY}

	if err := nilArrayTypeAssertion.Run(value); err != nil {
		return err
	}

	return nil
}
