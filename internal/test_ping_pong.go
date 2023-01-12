package internal

import (
	"fmt"
	testerutils "github.com/codecrafters-io/tester-utils"
	"github.com/go-redis/redis"
	"net"
	"time"
)

func testPingPongOnce(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	retries := 0
	var err error
	var conn net.Conn

	for {
		conn, err = net.Dial("tcp", "localhost:6379")
		if err != nil && retries > 15 {
			logger.Infof("All retries failed.")
			return err
		}

		if err != nil {
			// Don't print errors in the first second
			if retries > 2 {
				logger.Infof("Failed to connect to port 6379, retrying in 1s")
			}

			retries += 1
			time.Sleep(1000 * time.Millisecond)
		} else {
			break
		}
	}

	logger.Debugln("Connection established, sending PING command (*1\\r\\n$4\\r\\nping\\r\\n)")

	_, err = conn.Write([]byte("*1\r\n$4\r\nping\r\n"))
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	time.Sleep(100 * time.Millisecond) // Ensure we aren't reading partial responses

	logger.Debugln("Reading response...")

	var readBytes = make([]byte, 16)

	numberOfBytesRead, err := conn.Read(readBytes)
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	actual := string(readBytes[:numberOfBytesRead])
	expected1 := "+PONG\r\n"
	expected2 := "$4\r\nPONG\r\n"

	if actual != expected1 && actual != expected2 {
		return fmt.Errorf("expected response to be either %#v or %#v, got %#v", expected1, expected2, actual)
	}

	return nil
}

func testPingPongMultiple(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client := NewRedisClient()

	for i := 1; i <= 3; i++ {
		if err := runPing(logger, client, 1); err != nil {
			return err
		}
	}

	logger.Debugf("Success, closing connection...")
	client.Close()

	return nil
}

func testPingPongConcurrent(stageHarness *testerutils.StageHarness) error {
	b := NewRedisBinary(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	client1 := NewRedisClient()

	if err := runPing(logger, client1, 1); err != nil {
		return err
	}

	client2 := NewRedisClient()
	if err := runPing(logger, client2, 2); err != nil {
		return err
	}

	if err := runPing(logger, client1, 1); err != nil {
		return err
	}
	if err := runPing(logger, client1, 1); err != nil {
		return err
	}
	if err := runPing(logger, client2, 2); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 1)
	client1.Close()

	client3 := NewRedisClient()
	if err := runPing(logger, client3, 3); err != nil {
		return err
	}

	logger.Debugf("client-%d: Success, closing connection...", 2)
	client2.Close()
	logger.Debugf("client-%d: Success, closing connection...", 3)
	client3.Close()

	return nil
}

func runPing(logger *testerutils.Logger, client *redis.Client, clientNum int) error {
	logger.Debugf("client-%d: Sending ping command...", clientNum)
	pong, err := client.Ping().Result()
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}

	logger.Debugf("client-%d: Received response.", clientNum)

	if pong != "PONG" {
		return fmt.Errorf("client-%d: Expected \"PONG\", got %#v", clientNum, pong)
	}

	return nil
}
