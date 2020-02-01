package main

import (
	"fmt"
	"net"
)

func testMultipleClients(executable *Executable, logger *customLogger) error {
	b := NewRedisBinary(executable, logger)
	if err := b.Run(); err != nil {
		return err
	}
	defer b.Kill()

	logger.Debugf("Creating first connection")
	conn1, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}

	logger.Debugf("Creating second connection")
	conn2, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}

	logger.Debugf("Sending ping to connection 1")
	err = sendPing(conn1, logger)
	if err != nil {
		return err
	}

	logger.Debugf("Sending ping to connection 2")
	err = sendPing(conn2, logger)
	if err != nil {
		return err
	}

	logger.Debugf("Sending ping to connection 1 again")
	err = sendPing(conn1, logger)
	if err != nil {
		return err
	}

	logger.Debugf("Sending ping to connection 2 again")
	err = sendPing(conn2, logger)
	if err != nil {
		return err
	}

	return nil
}

func sendPing(conn net.Conn, logger *customLogger) error {
	tmp := make([]byte, 256)

	logger.Debugf("- Writing PING command")
	_, err := conn.Write([]byte("*1\r\n$4\r\nping\r\n"))
	if err != nil {
		return err
	}
	logger.Debugf("- Reading response")
	readlen, err := conn.Read(tmp)
	if err != nil {
		return err
	}

	expected := []byte("+PONG\r\n")
	actual := tmp[:readlen]
	if string(actual[:]) != string(expected[:]) {
		return fmt.Errorf(`expected %s, got %s

expected bytes: %v
received bytes: %v`, expected, actual, expected, actual)
	}

	return nil
}
