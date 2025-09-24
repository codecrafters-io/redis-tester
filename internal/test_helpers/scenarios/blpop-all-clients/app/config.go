package main

import (
	"flag"
)

type Config struct {
	Port int
}

func ParseConfig() (*Config, error) {
	port := flag.Int("port", 6379, "Port to listen on")

	flag.Parse()

	config := &Config{
		Port: *port,
	}

	return config, nil
}
