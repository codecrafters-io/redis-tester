package resp_assertions

import (
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type MessageReceivedAssertion struct {
	channel string
	message string
}

func NewMessageReceivedAssertion(channel string, message string) RESPAssertion {
	return MessageReceivedAssertion{
		channel: channel,
		message: message,
	}
}

func (c MessageReceivedAssertion) Run(value resp_value.Value) error {
	arrayAssertion := NewOrderedStringArrayAssertion([]string{"message", c.channel, c.message})
	return arrayAssertion.Run(value)
}
