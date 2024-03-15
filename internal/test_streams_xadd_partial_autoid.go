package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testStreamsXaddPartialAutoid(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient("localhost:6379")

	randomKey := testerutils_random.RandomWord()

	tests := []XADDTest{
		{streamKey: randomKey, id: "0-*", values: map[string]interface{}{"foo": "bar"}, expectedResponse: "0-1"},
		{streamKey: randomKey, id: "1-*", values: map[string]interface{}{"foo": "bar"}, expectedResponse: "1-0"},
		{streamKey: randomKey, id: "1-*", values: map[string]interface{}{"bar": "baz"}, expectedResponse: "1-1"},
	}

	for _, test := range tests {
		err := test.Run(client, logger)

		if err != nil {
			return err
		}
	}

	return nil
}
