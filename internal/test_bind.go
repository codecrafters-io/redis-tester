package internal

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"net"
	"time"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testBindToPort(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	logger.Infof("Connecting to port 6379...")

	retries := 0
	var err error
	for {
		_, err = net.Dial("tcp", "localhost:6379")
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
				logger.Infof("Failed to connect to port 6379, retrying in 1s")
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
