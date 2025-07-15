package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	Port      int
	ReplicaOf *ReplicaConfig
}

type ReplicaConfig struct {
	Host string
	Port int
}

func ParseConfig() (*Config, error) {
	port := flag.Int("port", 6379, "Port to listen on")
	replicaOf := flag.String("replicaof", "", "Replica configuration in format 'host port'")

	flag.Parse()

	config := &Config{
		Port: *port,
	}

	if *replicaOf != "" {
		parts := strings.Fields(*replicaOf)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid replicaof format: expected 'host port', got '%s'", *replicaOf)
		}

		host := parts[0]
		replicaPort := 6379 // default port

		replicaPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port in --replicaof")
		}

		config.ReplicaOf = &ReplicaConfig{
			Host: host,
			Port: replicaPort,
		}
	}

	return config, nil
}
