package internal

import (
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testStreamsXaddPartialAutoid(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient("localhost:6379")

	strings := [10]string{
		"hello",
		"world",
		"mangos",
		"apples",
		"oranges",
		"watermelons",
		"grapes",
		"pears",
		"horses",
		"elephants",
	}

	randomKey := strings[rand.Intn(10)]

	tests := []XADDTest{
		{streamKey: randomKey, id: "0-*", values: map[string]interface{}{"foo": "bar"}, expectedResponse: "0-1"},
		{streamKey: randomKey, id: "1-*", values: map[string]interface{}{"foo": "bar"}, expectedResponse: "1-0"},
		{streamKey: randomKey, id: "1-*", values: map[string]interface{}{"bar": "baz"}, expectedResponse: "1-1"},
	}

	for _, test := range tests {
		testXadd(client, logger, test)
	}

	return nil
}
