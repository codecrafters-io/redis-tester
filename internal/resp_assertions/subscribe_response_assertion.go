package resp_assertions

import (
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
	arrayAssertion := NewOrderedArrayAssertion([]RESPAssertion{
		NewBulkStringAssertion("subscribe"),
		NewBulkStringAssertion(c.ExpectedChannel),
		NewIntegerAssertion(c.ExpectedSubscribedCount),
	})

	return arrayAssertion.Run(value)
}
