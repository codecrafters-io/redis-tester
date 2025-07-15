package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	config, err := ParseConfig()
	if err != nil {
		fmt.Println("Failed to parse configuration:", err.Error())
		os.Exit(1)
	}

	server := NewRedisServer(config)

	if err := server.Start(); err != nil {
		fmt.Println("Failed to start server:", err.Error())
		os.Exit(1)
	}
}
