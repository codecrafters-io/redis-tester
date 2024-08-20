package internal

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/go-redis/redis"
)

type XADDTest struct {
	streamKey        string
	id               string
	values           map[string]interface{}
	expectedResponse string
	expectedError    string
}

func (t *XADDTest) Run(client *redis.Client, logger *logger.Logger) error {
	var values []string

	for key, value := range t.values {
		values = append(values, key, fmt.Sprintf("%v", value))
	}

	logger.Infof("$ redis-cli xadd %v %v %v", t.streamKey, t.id, strings.Join(values, " "))

	resp, err := client.XAdd(&redis.XAddArgs{
		Stream: t.streamKey,
		ID:     t.id,
		Values: t.values,
	}).Result()

	if err != nil && t.expectedError == "" {
		logFriendlyError(logger, err)
		return err
	}

	if err != nil && t.expectedError != "" {
		if err.Error() != t.expectedError {
			return fmt.Errorf("Expected %q, got %q", t.expectedError, err.Error())
		}

		logger.Successf("Received error: %q", err.Error())
		return nil
	}

	if resp != t.expectedResponse && t.expectedError != "" {
		logger.Infof("Received response: %q", resp)
		return fmt.Errorf("Expected an error as the response, got %q", resp)
	} else if resp != t.expectedResponse {
		logger.Infof("Received response: %q", resp)
		return fmt.Errorf("Expected %q, got %q", t.expectedResponse, resp)
	} else {
		logger.Successf("Received response: %q", resp)
	}

	return nil
}

func testStreamsXadd(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	randomKey := random.RandomWord()

	multiCommandTestCase := test_cases.MultiCommandTestCase{
		Commands: [][]string{
			{"XADD", randomKey, "0-1", "foo", "bar"},
			{"TYPE", randomKey},
		},
		Assertions: []resp_assertions.RESPAssertion{
			resp_assertions.NewStringAssertion("0-1"),
			resp_assertions.NewStringAssertion("stream"),
		},
	}

	return multiCommandTestCase.RunAll(client, logger)
}
