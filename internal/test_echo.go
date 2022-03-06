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
		"pans",
		"pots",
		"cats",
		"apples",
		"oranges",
		"this",
		"is",
		"random",
		"input",
		"dogs",
	}

	randomString := strings[rand.Intn(10)]
	resp, err := client.Echo(randomString).Result()
	if err != nil {
		logger.Errorf(err.Error())
		logFriendlyError(logger, err)
		return err
	}

	if resp != randomString {
		err := fmt.Errorf("Expected %s, got %s", randomString, resp)
		logger.Errorf(err.Error())
		return err
	}

	client.Close()

	return nil
}
