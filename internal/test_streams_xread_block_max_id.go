package internal

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsXreadBlockMaxID(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client1, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "redis-cli")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client1.Close()

	streamKey := testerutils_random.RandomWord()
	entryValue := testerutils_random.RandomInt(1, 100)

	xaddCommandTestCase := &test_cases.SendCommandTestCase{
		Command:                   "XADD",
		Args:                      []string{streamKey, "0-1", "temperature", strconv.Itoa(entryValue)},
		Assertion:                 resp_assertions.NewStringAssertion("0-1"),
		ShouldSkipUnreadDataCheck: true,
	}

	if err := xaddCommandTestCase.Run(client1, logger); err != nil {
		return err
	}

	xReadResult := make(chan error, 1)
	entryValue = testerutils_random.RandomInt(1, 100)

	xReadAssertion := resp_assertions.NewXReadResponseAssertion([]resp_assertions.StreamResponse{{
		Key: streamKey,
		Entries: []resp_assertions.StreamEntry{{
			Id:              "0-2",
			FieldValuePairs: [][]string{{"temperature", strconv.Itoa(entryValue)}},
		}},
	}})
	xReadTestCase := &test_cases.SendCommandTestCase{
		Command:                   "XREAD",
		Args:                      []string{"block", "0", "streams", streamKey, "$"},
		Assertion:                 xReadAssertion,
		ShouldSkipUnreadDataCheck: true,
	}
	xReadTestCase.PauseReadingResponse()

	go func() {
		err := xReadTestCase.Run(client1, logger)
		xReadResult <- err
	}()

	time.Sleep(1000 * time.Millisecond)

	client2, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "other-redis-cli")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client2.Close()

	xaddCommandTestCase = &test_cases.SendCommandTestCase{
		Command:                   "XADD",
		Args:                      []string{streamKey, "0-2", "temperature", strconv.Itoa(entryValue)},
		Assertion:                 resp_assertions.NewStringAssertion("0-2"),
		ShouldSkipUnreadDataCheck: true,
	}

	if err := xaddCommandTestCase.Run(client2, logger); err != nil {
		return err
	}

	// Ensure the responses of XAdd and XRead are fixed in the fixture
	xReadTestCase.ResumeReadingResponse()
	err = <-xReadResult
	if err != nil {
		return err
	}

	xreadCommandTestCase := &test_cases.SendCommandTestCase{
		Command:                   "XREAD",
		Args:                      []string{"block", "1000", "streams", streamKey, "$"},
		Assertion:                 resp_assertions.NewNilAssertion(),
		ShouldSkipUnreadDataCheck: false,
	}

	if err := xreadCommandTestCase.Run(client1, logger); err != nil {
		return err
	}
	return nil
}
