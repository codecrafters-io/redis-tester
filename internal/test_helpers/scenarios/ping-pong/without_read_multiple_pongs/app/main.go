package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	connection, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	// Simulate case where user doesn't read commands one by one
	resp := []byte("+PONG\r\n+PONG\r\n+PONG\r\n")
	_, err = connection.Write(resp)
	if err != nil {
		fmt.Println("Error writing resp: ", err.Error())
		os.Exit(1)
	}
}
