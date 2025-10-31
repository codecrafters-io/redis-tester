package resp_assertions

import (
	"fmt"
	"slices"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type ArrayIndexAssertionSpecification struct {
	Index                int
	Assertion            RESPAssertion
	PreAssertionHook     func()
	AssertionSuccessHook func()
}

type ArrayElementsAssertion struct {
	IndexAssertionSpecifications []ArrayIndexAssertionSpecification
}

func (a ArrayElementsAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	array := value.Array()

	// Validate indexes
	for _, indexAssertionSpecification := range a.IndexAssertionSpecifications {
		if indexAssertionSpecification.Index < 0 {
			panic("Codecrafters Internal Error - Index in IndexAssertionSpecification is negative")
		}

		if indexAssertionSpecification.Index >= len(array) {
			return fmt.Errorf("Expected array to have at least %d elements, got %d", indexAssertionSpecification.Index+1, len(array))
		}
	}

	// Sort the indexes so the assertion runs serially
	slices.SortFunc(a.IndexAssertionSpecifications, func(ias1, ias2 ArrayIndexAssertionSpecification) int {
		return ias1.Index - ias2.Index
	})

	for _, indexAssertionSpecification := range a.IndexAssertionSpecifications {
		if indexAssertionSpecification.PreAssertionHook != nil {
			indexAssertionSpecification.PreAssertionHook()
		}

		if err := indexAssertionSpecification.Assertion.Run(array[indexAssertionSpecification.Index]); err != nil {
			return err
		}

		if indexAssertionSpecification.AssertionSuccessHook != nil {
			indexAssertionSpecification.AssertionSuccessHook()
		}
	}

	return nil
}
