package main

import (
	"fmt"
	"net"
	"time"
)

func testBindToPort(executable *Executable, logger *customLogger) error {
	logger.Debugf("Running program")
	if err := executable.Start(); err != nil {
		return err
	}
	defer executable.Kill()
	defer fmt.Println("after test")

	logger.Debugf("Creating first connection")
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}

	logger.Debugf("Writing a PING command")
	_, err = conn.Write([]byte("*1\r\n$4\r\nping\r\n"))
	if err != nil {
		return err
	}

	return nil
}
