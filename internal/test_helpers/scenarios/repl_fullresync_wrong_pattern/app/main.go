package main

import (
	"fmt"
	"os"
)

func main() {
	config, err := ParseConfig()
	if err != nil {
		fmt.Println("Failed to parse configuration:", err.Error())
		os.Exit(1)
	}

	fmt.Println("This server will intentionally fail at ")

	server := NewRedisServer(config)

	if err := server.Start(); err != nil {
		fmt.Println("Failed to start server:", err.Error())
		os.Exit(1)
	}
}
