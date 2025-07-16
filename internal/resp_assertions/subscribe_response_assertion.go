package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type SubscribeResponseAssertion struct {
	ExpectedChannel         string
	ExpectedSubscribedCount int
}

func NewSubscribeResponseAssertion(channel string, subscribedCount int) RESPAssertion {
	return SubscribeResponseAssertion{
		ExpectedChannel:         channel,
		ExpectedSubscribedCount: subscribedCount,
	}
}

func (c SubscribeResponseAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	arrayAssertion := NewOrderedArrayAssertion([]RESPAssertion{
		NewStringAssertion("subscribe"),
		NewStringAssertion(c.ExpectedChannel),
		NewIntegerAssertion(c.ExpectedSubscribedCount),
	})

	return arrayAssertion.Run(value)
}
