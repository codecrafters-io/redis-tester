package test_cases

import (
	"slices"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type subscriber struct {
	Client   *instrumented_resp_connection.InstrumentedRespConnection
	Channels []string
}

type PubSubTestCase struct {
	subscribers []subscriber
}

func NewPubSubTestCase() *PubSubTestCase {
	return &PubSubTestCase{subscribers: make([]subscriber, 0)}
}

func (t *PubSubTestCase) AddSubscription(client *instrumented_resp_connection.InstrumentedRespConnection, channel string) *PubSubTestCase {
	for i := range t.subscribers {
		if t.subscribers[i].Client == client {
			t.subscribers[i].Channels = append(t.subscribers[i].Channels, channel)
			return t
		}
	}
	t.subscribers = append(t.subscribers, subscriber{
		Client:   client,
		Channels: []string{channel},
	})
	return t
}

func (t *PubSubTestCase) RunSubscribeFromAll(logger *logger.Logger) error {
	// send subscribe from all clients in deterministic order
	for _, subscriber := range t.subscribers {
		// We issue SUBSCRIBE separately because we haven't introduced subscribing multiple channels using a single subscribe
		subscribeTestCase := MultiCommandTestCase{}
		for i, chanName := range subscriber.Channels {
			subscribeTestCase.CommandWithAssertions = append(subscribeTestCase.CommandWithAssertions, CommandWithAssertion{
				Command:   []string{"SUBSCRIBE", chanName},
				Assertion: resp_assertions.NewSubscribeResponseAssertion(chanName, i+1),
			})
		}
		if err := subscribeTestCase.RunAll(subscriber.Client, logger); err != nil {
			return err
		}
	}

	return nil
}

func (t *PubSubTestCase) getSubscriberFromClient(client *instrumented_resp_connection.InstrumentedRespConnection) (subscriber subscriber, idx int) {
	for i, c := range t.subscribers {
		if c.Client == client {
			return c, i
		}
	}
	panic("Codecrafters Internal Error - client is not a valid subscriber")
}

func (t *PubSubTestCase) RunUnsubscribe(client *instrumented_resp_connection.InstrumentedRespConnection, channel string, logger *logger.Logger) error {
	subscriber, subscriberIdx := t.getSubscriberFromClient(client)
	newSubscribedCount := len(subscriber.Channels)
	channelIndex := slices.Index(subscriber.Channels, channel)
	if channelIndex != -1 {
		newSubscribedCount -= 1
	}

	unsubscribeTestCase := SendCommandTestCase{
		Command: "UNSUBSCRIBE",
		Args:    []string{channel},
		Assertion: resp_assertions.NewOrderedArrayAssertion([]resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("unsubscribe"),
			resp_assertions.NewStringAssertion(channel),
			resp_assertions.NewIntegerAssertion(newSubscribedCount),
		}),
	}

	if err := unsubscribeTestCase.Run(client, logger); err != nil {
		return err
	}

	// remove the channel from the subscriber only if it was previously present
	if channelIndex != -1 {
		t.subscribers[subscriberIdx].Channels = append(subscriber.Channels[:channelIndex], subscriber.Channels[channelIndex+1:]...)
	}
	return nil
}

func (t *PubSubTestCase) RunPublishWithoutMessageAssertion(channel string, message string, publisher *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	subscriberCount := 0
	for _, subscriber := range t.subscribers {
		for _, ch := range subscriber.Channels {
			if ch == channel {
				subscriberCount++
			}
		}
	}
	publishTestCase := SendCommandTestCase{
		Command:   "PUBLISH",
		Args:      []string{channel, message},
		Assertion: resp_assertions.NewIntegerAssertion(subscriberCount),
	}

	return publishTestCase.Run(publisher, logger)
}

func (t *PubSubTestCase) assertMessages(channel string, message string, logger *logger.Logger) error {
	for _, subscriber := range t.subscribers {
		isSubscribedToChannel := slices.Contains(subscriber.Channels, channel)
		if isSubscribedToChannel {
			subscriber.Client.GetLogger().Infof("Expecting published message")
			receiveTestCase := ReceiveValueTestCase{
				Assertion: resp_assertions.NewMessageReceivedAssertion(channel, message),
			}
			if err := receiveTestCase.Run(subscriber.Client, logger); err != nil {
				return err
			}
		} else {
			noResponseTestCase := NoResponseTestCase{}
			if err := noResponseTestCase.Run(subscriber.Client); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *PubSubTestCase) RunPublish(channel string, message string, publisher *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	if err := t.RunPublishWithoutMessageAssertion(channel, message, publisher, logger); err != nil {
		return err
	}
	return t.assertMessages(channel, message, logger)
}
