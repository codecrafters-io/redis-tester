package internal

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	testerutils "github.com/codecrafters-io/tester-utils"
)

func testReplBindToCustomPort(stageHarness *testerutils.StageHarness) error {
	randomPortOffset := rand.Intn(11)
	port := 6380 + randomPortOffset

	b := NewRedisBinary(stageHarness)

	b.args = []string{
		"--port", strconv.Itoa(port),
	}

	if err := b.Run(); err != nil {
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
