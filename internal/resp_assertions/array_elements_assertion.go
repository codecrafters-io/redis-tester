package resp_assertions

import (
	"fmt"
	"slices"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ArrayIndexAssertionSpecification struct {
	ArrayElementAssertion ArrayElementAssertion
	PreAssertionHook      func()
	AssertionSuccessHook  func()
}

type ArrayElementsAssertion struct {
	ArrayElementAssertionSpecification []ArrayIndexAssertionSpecification
}

func (a ArrayElementsAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	// Sort the indexes so the assertion runs serially
	slices.SortFunc(a.ArrayElementAssertionSpecification, func(aea1, aea2 ArrayIndexAssertionSpecification) int {
		return aea1.ArrayElementAssertion.Index - aea2.ArrayElementAssertion.Index
	})

	multiAssertion := MultiAssertion{}

	for _, arrayElementAssertionSpecification := range a.ArrayElementAssertionSpecification {
		multiAssertion.AssertionSpecifications = append(multiAssertion.AssertionSpecifications, AssertionSpecification{
			Assertion:            arrayElementAssertionSpecification.ArrayElementAssertion,
			PreAssertionHook:     arrayElementAssertionSpecification.PreAssertionHook,
			AssertionSuccessHook: arrayElementAssertionSpecification.AssertionSuccessHook,
		})
	}

	return multiAssertion.Run(value)
}
