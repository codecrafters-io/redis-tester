package resp_assertions

import (
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type NilAssertion struct{}

func NewNilAssertion() RESPAssertion {
	return NilAssertion{}
}

func (a NilAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.NIL}
	return dataTypeAssertion.Run(value)
}
