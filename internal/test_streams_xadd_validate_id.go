package internal

import (
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testStreamsXaddValidateId(stageHarness *testerutils.StageHarness) error {
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
		{streamKey: randomKey, id: "1-1", values: map[string]interface{}{"foo": "bar"}, expectedResponse: "1-1", expectedError: ""},
		{streamKey: randomKey, id: "1-2", values: map[string]interface{}{"bar": "baz"}, expectedResponse: "1-2", expectedError: ""},
		{streamKey: randomKey, id: "1-2", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
		{streamKey: randomKey, id: "0-3", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
		{streamKey: randomKey, id: "0-0", values: map[string]interface{}{"baz": "foo"}, expectedResponse: "", expectedError: "ERR The ID specified in XADD must be greater than 0-0"},
	}

	for _, test := range tests {
		testXadd(client, logger, test)
	}

	return nil
}
