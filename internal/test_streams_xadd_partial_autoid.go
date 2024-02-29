package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func testStreamsXaddPartialAutoid(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
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
		err := testXadd(client, logger, test)

		if err != nil {
			return err
		}
	}

	return nil
}
