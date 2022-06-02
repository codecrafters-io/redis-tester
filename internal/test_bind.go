package internal

import (
	"fmt"
	"net"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testBindToPort(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

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

	logger.Debugf("Connection successful")
	return nil
}
