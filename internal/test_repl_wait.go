package internal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"

	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

type WaitTest struct {
	// WriteCommand is the command we'll issue to the master
	WriteCommand []string

	// WaitReplicaCount is the number of replicas we'll specify in the WAIT command
	WaitReplicaCount int

	// WaitTimeoutMilli is the timeout we'll specify in the WAIT command
	WaitTimeoutMilli int

	// ActualNumberOfAcks is the number of ACKs we'll send back to the master
	ActualNumberOfAcks int

	// ShouldVerifyTimeout is a flag to verify if the WAIT command returned after the timeout
	ShouldVerifyTimeout bool

	// Logger is the logger to use for this test
	Logger *logger.Logger
}

// In this stage, we:
//  1. Boot the user's code as a Redis master.
//  2. Spawn multiple replicas and have each perform a handshake with the master.
//  3. Connect to Master, and execute RunWaitTest
//  4. RunWaitTest :
//     4.1. Issue a write command to the master
//     4.2. Issue a WAIT command with WaitReplicaCount as the expected number of replicas
//     4.3. Read propagated command on replicas + respond to subset of GETACKs
//     4.4. Assert response of WAIT command is ActualNumberOfAcks
//     4.5. Assert that the WAIT command returned after the timeout
func testWait(stageHarness *test_case_harness.TestCaseHarness) error {
	deleteRDBfile()

	// Step 1: Boot the user's code as a Redis master.
	master := redis_executable.NewRedisExecutable(stageHarness)
	if err := master.Run("--port", "6379"); err != nil {
		return err
	}

	logger := stageHarness.Logger
	defer logger.ResetSecondaryPrefixes()

	// Step 2: Spawn multiple replicas and have each perform a handshake
	replicaCount := testerutils_random.RandomInt(3, 5)
	logger.Infof("Proceeding to create %v replicas.", replicaCount)

	replicas, err := SpawnReplicas(replicaCount, stageHarness, logger, "localhost:6379")
	if err != nil {
		return err
	}
	for _, replica := range replicas {
		defer replica.Close()
	}

	// Step 3: Connect to master
	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	logger.UpdateLastSecondaryPrefix("test")
	client.UpdateBaseLogger(logger)
	for _, r := range replicas {
		r.UpdateBaseLogger(logger)
	}

	if err = RunWaitTest(client, replicas, WaitTest{
		WriteCommand:        []string{"SET", "foo", "123"},
		WaitReplicaCount:    1,
		ActualNumberOfAcks:  1,
		WaitTimeoutMilli:    500,
		ShouldVerifyTimeout: false,
		Logger:              logger,
	}); err != nil {
		return err
	}

	logger.Successf("Passed first WAIT test.")

	waitCommandReplicaSubsetCount := testerutils_random.RandomInt(2, replicaCount) + 1
	if err = RunWaitTest(client, replicas, WaitTest{
		WriteCommand:        []string{"SET", "baz", "789"},
		WaitReplicaCount:    waitCommandReplicaSubsetCount,
		ActualNumberOfAcks:  waitCommandReplicaSubsetCount - 1,
		WaitTimeoutMilli:    2000,
		ShouldVerifyTimeout: true,
		Logger:              logger,
	}); err != nil {
		return err
	}

	return nil
}

func consumeReplicationStreamAndSendAcks(replicas []*instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger, acksSentByReplicaSubsetCount int, command []string) error {
	var err error
	for j := 0; j < len(replicas); j++ {
		replica := replicas[j]
		logger.Infof("Testing Replica: %s", replica.GetIdentifier())

		replicaLogger := replica.GetLogger()
		replicaLogger.Infof("Expecting \"%s\" to be propagated", strings.Join(command, " "))

		receiveCommandTestCase := &test_cases.ReceiveValueTestCase{
			Assertion:                 resp_assertions.NewCommandAssertion(command[0], command[1:]...),
			ShouldSkipUnreadDataCheck: true,
		}

		err = receiveCommandTestCase.Run(replica, logger)

		if err != nil {
			// Redis sends a SELECT command, but we don't expect it from users.
			// If the first command is a SELECT command, we'll re-run the test case to test the next command instead
			if IsSelectCommand(receiveCommandTestCase.ActualValue) {
				err := receiveCommandTestCase.Run(replica, logger)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		replicaLogger.Infof("Expecting \"REPLCONF GETACK *\" from Master")

		receiveGetackCommandTestCase := &test_cases.ReceiveValueTestCase{
			Assertion:                 resp_assertions.NewCommandAssertion("REPLCONF", "GETACK", "*"),
			ShouldSkipUnreadDataCheck: false,
		}
		if err = receiveGetackCommandTestCase.Run(replica, logger); err != nil {
			return err
		}

		if j < acksSentByReplicaSubsetCount {
			replicaLogger.Debugf("Sending ACK to Master")
			// Remove GETACK command bytes from offset before sending ACK.
			if err := replica.SendCommand("REPLCONF", []string{"ACK", strconv.Itoa(replica.ReceivedBytesCount - replica.LastValueBytesCount)}...); err != nil {
				return err
			}
		} else {
			replicaLogger.Debugf("Not sending ACK to Master")
		}
	}
	return err
}

func RunWaitTest(client *instrumented_resp_connection.InstrumentedRespConnection, replicas []*instrumented_resp_connection.InstrumentedRespConnection, waitTest WaitTest) (err error) {
	// Step 1: Issue a write command
	setCommandTestCase := test_cases.SendCommandTestCase{
		Command:   waitTest.WriteCommand[0],
		Args:      waitTest.WriteCommand[1:],
		Assertion: resp_assertions.NewStringAssertion("OK"),
	}
	if err := setCommandTestCase.Run(client, waitTest.Logger); err != nil {
		return err
	}

	// Step 2: Issue a WAIT command with a subset as the expected number of replicas
	startTimeMilli := time.Now().UnixMilli()
	if err := client.SendCommand("WAIT", []string{strconv.Itoa(waitTest.WaitReplicaCount), strconv.Itoa(waitTest.WaitTimeoutMilli)}...); err != nil {
		return err
	}

	// Step 3: Read propagated command on replicas + respond to subset of GETACKs
	// We then assert that across all the replicas we receive the SET commands in order
	err = consumeReplicationStreamAndSendAcks(replicas, waitTest.Logger, waitTest.ActualNumberOfAcks, waitTest.WriteCommand)
	if err != nil {
		return err
	}

	// Step 4: Assert response of WAIT command is replicaAcksCount
	value, err := client.ReadValueWithTimeout(4 * time.Second)
	if err != nil {
		return err
	}

	if err := resp_assertions.NewIntegerAssertion(waitTest.ActualNumberOfAcks).Run(value); err != nil {
		return err
	}

	endTimeMilli := time.Now().UnixMilli()

	// Step 5: If shouldVerifyTimeout is true : Assert that the WAIT command returned after the timeout
	if waitTest.ShouldVerifyTimeout {
		threshold := 500 // ms
		elapsedTimeMilli := endTimeMilli - startTimeMilli
		waitTest.Logger.Infof("%s", fmt.Sprintf("WAIT command returned after %v ms", elapsedTimeMilli))
		if math.Abs(float64(elapsedTimeMilli-int64(waitTest.WaitTimeoutMilli))) > float64(threshold) {
			return fmt.Errorf("Expected WAIT to return exactly after %v ms timeout elapsed.", 1000)
		}
	}

	return nil
}
