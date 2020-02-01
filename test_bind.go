package main

import (
	"net"
)

func testBindToPort(executable *Executable, logger *customLogger) error {
	b := NewRedisBinary(executable, logger)
	if err := b.Run(); err != nil {
		return err
	}
	defer b.Kill()

	logger.Debugf("Creating first connection")
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
