package resp_assertions

import (
	"fmt"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type SubscribeResponseAssertion struct {
	channel         string
	subscribedCount int
}

func NewSubscribeResponseAssertion(channel string, subscribedCount int) RESPAssertion {
	return SubscribeResponseAssertion{
		channel:         channel,
		subscribedCount: subscribedCount,
	}
}

func (c SubscribeResponseAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.ARRAY {
		return fmt.Errorf("Expected array, got %s", value.Type)
	}

	respArray := value.Array()

	if len(respArray) != 3 {
		return fmt.Errorf("Expected array length: 3. Got array length: %d", len(respArray))
	}

	subscribeLiteral := respArray[0]
	channelName := respArray[1]
	subscribeCount := respArray[2]

	subscribeAssertion := NewStringAssertion("subscribe")
	if err := subscribeAssertion.Run(subscribeLiteral); err != nil {
		return err
	}

	channelNameAssertion := NewStringAssertion(c.channel)
	if err := channelNameAssertion.Run(channelName); err != nil {
		return err
	}

	subscribeCountAssertion := NewIntegerAssertion(c.subscribedCount)
	if err := subscribeCountAssertion.Run(subscribeCount); err != nil {
		return err
	}

	return nil
}
