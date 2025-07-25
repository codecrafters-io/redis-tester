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

type SubscriberGroupTestCase struct {
	subscribers []subscriber
}

func (t *SubscriberGroupTestCase) AddSubscription(client *instrumented_resp_connection.InstrumentedRespConnection, channel string) *SubscriberGroupTestCase {
	for i, subscriber := range t.subscribers {
		if subscriber.Client == client {
			// prevent duplicates
			if !slices.Contains(subscriber.Channels, channel) {
				t.subscribers[i].Channels = append(subscriber.Channels, channel)
			}
			return t
		}
	}
	t.subscribers = append(t.subscribers, subscriber{
		Client:   client,
		Channels: []string{channel},
	})
	return t
}

func (t *SubscriberGroupTestCase) RemoveSubscription(client *instrumented_resp_connection.InstrumentedRespConnection, channel string) *SubscriberGroupTestCase {
	for idx, s := range t.subscribers {
		if s.Client == client {
			channelIndex := slices.Index(s.Channels, channel)
			if channelIndex == -1 {
				// ignore if the client was not subscribed to the channel
				return t
			}
			// exclude the channel
			t.subscribers[idx].Channels = append(s.Channels[:channelIndex], s.Channels[channelIndex+1:]...)
			// remove the subscriber if no channels left
			if len(t.subscribers[idx].Channels) == 0 {
				t.subscribers = append(t.subscribers[:idx], t.subscribers[idx+1:]...)
			}
			return t
		}
	}
	// ignore if the client was not subscribed to anything
	return t
}

func (t *SubscriberGroupTestCase) GetSubscriberCount(channel string) int {
	total := 0
	for _, subscriber := range t.subscribers {
		if slices.Contains(subscriber.Channels, channel) {
			total++
		}
	}
	return total
}

func (t *SubscriberGroupTestCase) RunSubscribe(logger *logger.Logger) error {
	for _, subscriber := range t.subscribers {
		subscribeTestCase := MultiCommandTestCase{}
		for i, channel := range subscriber.Channels {
			subscribeTestCase.CommandWithAssertions = append(subscribeTestCase.CommandWithAssertions, CommandWithAssertion{
				Command:   []string{"SUBSCRIBE", channel},
				Assertion: resp_assertions.NewSubscribeResponseAssertion(channel, i+1),
			})
		}
		if err := subscribeTestCase.RunAll(subscriber.Client, logger); err != nil {
			return err
		}
	}
	return nil
}

func (t *SubscriberGroupTestCase) RunAssertionForPublishedMessage(channel string, message string, logger *logger.Logger) error {
	for _, subscriber := range t.subscribers {
		if slices.Contains(subscriber.Channels, channel) {
			subscriber.Client.GetLogger().Infof("Expecting published message: %q", message)
			receiveTestCase := ReceiveValueTestCase{
				Assertion: resp_assertions.NewPublishedMessageAssertion(channel, message),
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
