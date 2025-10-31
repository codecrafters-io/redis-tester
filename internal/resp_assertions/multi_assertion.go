package resp_assertions

import (
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type AssertionSpecification struct {
	Assertion            RESPAssertion
	PreAssertionHook     func()
	AssertionSuccessHook func()
}

type MultiAssertion struct {
	AssertionSpecifications []AssertionSpecification
}

func (a MultiAssertion) Run(value resp_value.Value) error {
	for _, assertionSpecification := range a.AssertionSpecifications {
		if assertionSpecification.PreAssertionHook != nil {
			assertionSpecification.PreAssertionHook()
		}

		if err := assertionSpecification.Assertion.Run(value); err != nil {
			return err
		}

		if assertionSpecification.AssertionSuccessHook != nil {
			assertionSpecification.AssertionSuccessHook()
		}

	}

	return nil
}
