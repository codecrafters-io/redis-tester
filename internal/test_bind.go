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
		if err != nil && retries > 20 {
			logger.Infof("All retries failed.")
			return err
		}

		if err != nil {
			if b.HasExited() {
				return fmt.Errorf("process has terminated, expected server to a be a long-running process")
			}

			logger.Infof("Failed to connect to port 6379, retrying in 500ms")
			retries += 1
			time.Sleep(500 * time.Millisecond)
		} else {
			break
		}
	}

	logger.Debugf("Connection successful")
	return nil
}
