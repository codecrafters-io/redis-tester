package main

import "fmt"

import "net"

func runStage2(logger *customLogger) error {
	conn1, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}

	conn2, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}

	logger.Debugf("Sending ping to connection 1")
	err = sendPing(conn1)
	if err != nil {
		return err
	}

	err = sendPing(conn2)
	if err != nil {
		return err
	}

	// 	err = sendPing(conn1)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = sendPing(conn2)
	// 	if err != nil {
	// 		return err
	// 	}

	return nil
}

func sendPing(conn net.Conn) error {
	tmp := make([]byte, 256)

	_, err := conn.Write([]byte("*1\r\n$4\r\nping\r\n"))
	if err != nil {
		return err
	}
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
