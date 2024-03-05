package main

import (
	"os"
	"strings"

	"github.com/codecrafters-io/redis-tester/internal"
)

func main() {
	os.Exit(internal.RunCLI(envMap()))
}

func envMap() map[string]string {
	result := make(map[string]string)
	for _, keyVal := range os.Environ() {
		split := strings.SplitN(keyVal, "=", 2)
		key, val := split[0], split[1]
		result[key] = val
	}

	return result
}
