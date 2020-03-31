package main

import (
	"github.com/codecrafters-io/redis-tester/internal"
	"os"
	"strings"
)

func main() {
	env := envMap()

	if _, ok := env["CODECRAFTERS_SUBMISSION_DIR"]; !ok {
		env["CODECRAFTERS_SUBMISSION_DIR"] = "/app"
	}

	env["CODECRAFTERS_CURRENT_STAGE_SLUG"] = "init"

	os.Exit(internal.RunCLI(env))
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
