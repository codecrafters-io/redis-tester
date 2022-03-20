package internal

import (
	"fmt"
	"math/rand"

	testerutils "github.com/codecrafters-io/tester-utils"
)

// Tests 'ECHO'
func testEcho(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := NewRedisClient()

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

	randomString := strings[rand.Intn(10)]
	logger.Debugf("Sending command: echo %s", randomString)
	resp, err := client.Echo(randomString).Result()
	if err != nil {
		logger.Errorf(err.Error())
		logFriendlyError(logger, err)
		return err
	}

	if resp != randomString {
		return fmt.Errorf("Expected %#v, got %#v", randomString, resp)
	}

	client.Close()

	return nil
}
