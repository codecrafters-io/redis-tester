package main

import "net"

func testBindToPort(logger *customLogger) error {
	logger.Debugf("Creating first connection")
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte("*1\r\n$4\r\nping\r\n"))
	if err != nil {
		return err
	}

	return nil
}
