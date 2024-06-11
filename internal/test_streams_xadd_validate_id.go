package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsXaddValidateID(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()

	tests := []XADDTest{
		{streamKey: randomKey, id: "100-1", values: map[string]interface{}{"foo": "bar"}, expectedResponse: "1-1", expectedError: ""},
		{streamKey: randomKey, id: "100-2", values: map[string]interface{}{"bar": "baz"}, expectedResponse: "1-2", expectedError: ""},
		{streamKey: randomKey, id: "100-2", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
		// This will catch an incorrect implementation using lexicographical comparison
		{streamKey: randomKey, id: "11-2", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
		{streamKey: randomKey, id: "0-3", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
		{streamKey: randomKey, id: "0-0", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD must be greater than 0-0"},
	}

	for _, test := range tests {
		err := test.Run(client, logger)

		if err != nil {
			return err
		}
	}

	return nil
}
