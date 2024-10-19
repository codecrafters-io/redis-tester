package test_cases

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/tester-utils/logger"
)

type BindTestCase struct {
	Port    int
	Retries int
}

func (t BindTestCase) Run(executable *redis_executable.RedisExecutable, logger *logger.Logger) error {
	logger.Infof("Connecting to port %d...", t.Port)

	retries := 0
	var err error
	address := "localhost:" + strconv.Itoa(t.Port)
	for {
		_, err = net.Dial("tcp", address)
		if err != nil && retries > t.Retries {
			logger.Infof("All retries failed.")
			return err
		}

		if err != nil {
			if executable.HasExited() {
				// We don't need to mention that the user's program exited or is expected to be a long-running process as
				// this could be confusing in early stages where the user is expected to only handle a single request from
				// a single client.
				//
				// Let's just exit early and not retry if this happens.
				return fmt.Errorf("Failed to connect to port %d.", t.Port)
			}

			// Don't print errors in the first second
			if retries > 2 {
				logger.Infof("Failed to connect to port %d, retrying in 1s", t.Port)
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
