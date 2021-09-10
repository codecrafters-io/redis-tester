package internal

import (
	"fmt"
	"math/rand"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
)

// Tests 'ECHO'
func testEcho(stageHarness testerutils.StageHarness) error {
	logger := stageHarness.Logger

	b := NewRedisBinary(stageHarness.Executable, logger)
	if err := b.Run(); err != nil {
		logger.Errorf(err.Error())
		return err
	}
	defer b.Kill()

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
