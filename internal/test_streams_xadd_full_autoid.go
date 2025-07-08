package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsXaddFullAutoid(stageHarness *test_case_harness.TestCaseHarness) error {
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

	streamKey := random.RandomWord()

	commandTestCase := &test_cases.SendCommandTestCase{
		Command:                   "XADD",
		Args:                      []string{streamKey, "*", "foo", "bar"},
		Assertion:                 resp_assertions.NewNoopAssertion(),
		ShouldSkipUnreadDataCheck: true,
	}

	if err := commandTestCase.Run(client, logger); err != nil {
		return err
	}

	responseValue := commandTestCase.ReceivedResponse

	if responseValue.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected bulk string, got %s", responseValue.Type)
	}

	parts := strings.Split(responseValue.String(), "-")

	if len(parts) != 2 {
		return fmt.Errorf("Expected a string in the form \"<millisecondsTime>-<sequenceNumber>\", got %q", responseValue)
	}

	timeStr, sequenceNumber := parts[0], parts[1]
	timeInt64, _ := strconv.ParseInt(timeStr, 10, 64)
	now := time.Now().Unix() * 1000
	oneSecondAgo := now - 1000
	oneSecondLater := now + 1000

	if len(timeStr) != 13 {
		return fmt.Errorf("Expected the first part of the ID to be a unix timestamp (%d characters), got %d characters", len(strconv.FormatInt(now, 10)), len(timeStr))
	} else if timeInt64 <= oneSecondAgo || timeInt64 >= oneSecondLater {
		return fmt.Errorf("Expected the first part of the ID to be a valid unix timestamp, got %q", timeStr)
	} else {
		logger.Successf("The first part of the ID is a valid unix milliseconds timestamp")
	}

	if sequenceNumber != "0" {
		return fmt.Errorf("Expected the second part of the ID to be a sequence number with a value of \"0\", got %q", sequenceNumber)
	} else {
		logger.Successf("The second part of the ID is a valid sequence number")
	}

	return nil
}
