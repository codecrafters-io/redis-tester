package internal

import (
	"fmt"
	"math/rand"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

// Tests 'ECHO'
func testEcho(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		DialTimeout: 30 * time.Second,
	})

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
	logger.Debugf("Sending command: ECHO %s", randomString)
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
