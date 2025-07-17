package resp_assertions

import (
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type PublishedMessageAssertion struct {
	ExpectedChannel string
	ExpectedMessage string
}

func NewPublishedMessageAssertion(channel string, message string) RESPAssertion {
	return PublishedMessageAssertion{
		ExpectedChannel: channel,
		ExpectedMessage: message,
	}
}

func (c PublishedMessageAssertion) Run(value resp_value.Value) error {
	arrayAssertion := NewOrderedStringArrayAssertion([]string{"message", c.ExpectedChannel, c.ExpectedMessage})
	return arrayAssertion.Run(value)
}
