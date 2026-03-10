package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingUnwatch(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clients, err := SpawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, c := range clients {
		defer c.Close()
	}

	keys := testerutils_random.RandomWords(2)
	initialValues := testerutils_random.RandomInts(1, 100, 2)

	key1, key2 := keys[0], keys[1]
	key1InitialValue, key2InitialValue := initialValues[0], initialValues[1]

	// Client 1: Set initial values for both keys
	setKeysTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", key1, strconv.Itoa(key1InitialValue)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", key2, strconv.Itoa(key2InitialValue)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
	}
	if err := setKeysTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Watch both keys
	watchTestCase := test_cases.WatchTestCase{Keys: []string{key1, key2}}

	if err := watchTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Modify key1 (a watched key)
	key1ValueSetByClient2 := testerutils_random.RandomInt(200, 400)

	modifyWatchedKeyTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{key1, strconv.Itoa(key1ValueSetByClient2)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := modifyWatchedKeyTestCase.Run(clients[1], logger); err != nil {
		return err
	}

	// Client 1: UNWATCH to clear all watched keys
	unwatchTestCase := test_cases.SendCommandTestCase{
		Command:   "UNWATCH",
		Args:      []string{},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := unwatchTestCase.Run(clients[0], logger); err != nil {
		return err
	}

	// Client 1: Run a transaction updating both keys
	key1ValueSetByClient1InTransaction := testerutils_random.RandomInt(500, 700)
	key2ValueSetByClient1InTransaction := testerutils_random.RandomInt(700, 1000)

	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", key1, strconv.Itoa(key1ValueSetByClient1InTransaction)},
			{"SET", key2, strconv.Itoa(key2ValueSetByClient1InTransaction)},
		},
		// Transaction should succeed since UNWATCH cleared the watched keys
		ExpectedResponseArray: []resp_assertions.RESPAssertion{
			resp_assertions.NewSimpleStringAssertion("OK"),
			resp_assertions.NewSimpleStringAssertion("OK"),
		},
	}

	if err := transactionTestCase.RunAll(clients[0], logger); err != nil {
		return err
	}

	// Client 2: Verify that the transaction was executed
	return (&test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"GET", key1},
				Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(key1ValueSetByClient1InTransaction)),
			},
			{
				Command:   []string{"GET", key2},
				Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(key2ValueSetByClient1InTransaction)),
			},
		},
	}).RunAll(clients[1], logger)
}
