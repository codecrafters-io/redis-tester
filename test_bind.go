package main

import (
	"net"
	"time"
)

func testBindToPort(executable *Executable, logger *customLogger) error {
	b := NewRedisBinary(executable, logger)
	if err := b.Run(); err != nil {
		return err
	}
	defer b.Kill()

	logger.Debugf("Creating first connection")
	retries := 0
	var err error
	for {
		_, err = net.Dial("tcp", "localhost:6379")
		if err != nil && retries > 5 {
			logger.Debugf("All retries failed.")
			return err
		}

		if err != nil {
			logger.Debugf("Failed, retrying in one second")
			retries += 1
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	logger.Debugf("Connection successful")
	return nil
}
