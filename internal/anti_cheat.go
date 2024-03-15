package internal

import (
	"fmt"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"net"
	"strings"
	"time"

	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func antiCheatTest(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	conn, err := NewRedisConn("", "localhost:6379")
	if err != nil {
		logger.Debugf("Error connecting to TCP server: %v", err)
		return err
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		return fmt.Errorf("Error setting read deadline: %w", err)
	}

	client := NewFakeRedisClient(conn, logger)
	if err := client.Send([]string{"MEMORY", "DOCTOR"}); err != nil {
		return fmt.Errorf("Error sending command to Redis: %w", err)
	}

	actualMessage, err := client.readRespString()
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil // Read timed out. No data received from client.
		}
		return nil
	}

	// All the answers for MEMORY DOCTOR include the string "sam" in them.
	if strings.Contains(strings.ToLower(actualMessage), "sam") {
		logger.Criticalf("anti-cheat (ac1) failed.")
		logger.Criticalf("Are you sure you aren't running this against the actual Redis?")
		return fmt.Errorf("anti-cheat (ac1) failed")
	} else {
		return nil
	}
}
