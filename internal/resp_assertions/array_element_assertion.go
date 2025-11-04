package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ArrayElementAssertion struct {
	Index     int
	Assertion RESPAssertion
}

func (a ArrayElementAssertion) Run(value resp_value.Value) error {
	arrayTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := arrayTypeAssertion.Run(value); err != nil {
		return err
	}

	array := value.Array()

	if a.Index < 0 {
		panic("Codecrafters Internal Error - Index in ArrayElementAssertion is negative")
	}

	if a.Index >= len(array) {
		return fmt.Errorf("Expected the array length to be at least %d, got %d",
			a.Index+1, len(array))
	}

	element := array[a.Index]

	return a.Assertion.Run(element)
}
