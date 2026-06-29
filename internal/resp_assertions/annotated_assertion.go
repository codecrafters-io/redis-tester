package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type AnnotatedAssertion struct {
	Annotation string
	Assertion  RESPAssertion
}

func NewAnnotatedAssertion(annotation string, assertion RESPAssertion) RESPAssertion {
	return AnnotatedAssertion{
		Annotation: annotation,
		Assertion:  assertion,
	}
}

func (a AnnotatedAssertion) Run(value resp_value.Value) error {
	if err := a.Assertion.Run(value); err != nil {
		return fmt.Errorf("%s: %w", a.Annotation, err)
	}

	return nil
}
