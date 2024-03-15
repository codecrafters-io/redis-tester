package internal

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testReplBindToCustomPort(stageHarness *test_case_harness.TestCaseHarness) error {
	port := testerutils_random.RandomInt(6380, 6390)

	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run("--port", strconv.Itoa(port)); err != nil {
		return err
	}

	logger := stageHarness.Logger

	logger.Infof("Connecting to port %d...", port)
	retries := 0
	var err error
	address := "localhost:" + strconv.Itoa(port)

	for {
		_, err = net.Dial("tcp", address)
		if err != nil && retries > 15 {
			logger.Infof("All retries failed.")
			return err
		}

		if err != nil {
			if b.HasExited() {
				return fmt.Errorf("Looks like your program has terminated. A redis server is expected to be a long-running process.")
			}

			// Don't print errors in the first second
			if retries > 2 {
				logger.Infof("Failed to connect to port %d, retrying in 1s", port)
			}

			retries += 1
			time.Sleep(1000 * time.Millisecond)
		} else {
			break
		}
	}

	logger.Debugln("Connection successful")
	return nil
}
