package resp_assertions

import resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"

type CompositeAssertion struct {
	Assertions []RESPAssertion
}

func NewCompositeAssertion(assertions ...RESPAssertion) *CompositeAssertion {
	return &CompositeAssertion{
		Assertions: assertions,
	}
}

func (a CompositeAssertion) Run(value resp_value.Value) RESPAssertionResult {
	var SuccessMessages, ErrorMessages []string

	for _, assertion := range a.Assertions {
		assertionResult := assertion.Run(value)
		if assertionResult.IsFailure() {
			ErrorMessages = append(ErrorMessages, assertionResult.ErrorMessages...)
		} else {
			SuccessMessages = append(SuccessMessages, assertionResult.SuccessMessages...)
		}
	}

	return RESPAssertionResult{
		SuccessMessages: SuccessMessages,
		ErrorMessages:   ErrorMessages,
	}
}
