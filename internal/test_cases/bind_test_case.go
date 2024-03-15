package test_cases

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/tester-utils/logger"
	"net"
	"strconv"
	"time"
)

type BindTestCase struct {
	Port    int
	Retries int
}

func (t BindTestCase) Run(executable *redis_executable.RedisExecutable, logger *logger.Logger) error {
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
				return fmt.Errorf("Looks like your program has terminated. A redis server is expected to be a long-running process.")
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
