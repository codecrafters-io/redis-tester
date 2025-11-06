package resp_assertions

import (
	"slices"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ArrayElementAssertionSpecification struct {
	ArrayElementAssertion ArrayElementAssertion
	PreAssertionHook      func()
	AssertionSuccessHook  func()
}

type ArrayElementsAssertion struct {
	ArrayElementAssertionSpecification []ArrayElementAssertionSpecification
}

func (a ArrayElementsAssertion) Run(value resp_value.Value) error {
	dataTypeAssertion := DataTypeAssertion{ExpectedType: resp_value.ARRAY}

	if err := dataTypeAssertion.Run(value); err != nil {
		return err
	}

	// Sort the indexes so the assertion runs serially
	slices.SortFunc(a.ArrayElementAssertionSpecification, func(aea1, aea2 ArrayElementAssertionSpecification) int {
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
