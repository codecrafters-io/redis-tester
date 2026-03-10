package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testOptimisticLockingTrackingKeyModification(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)

	if err := b.Run(); err != nil {
		return err
	}

	allKeys := testerutils_random.RandomWords(4)

	watchedKey, unwatchedKey := allKeys[0], allKeys[1]
	if err := testOptimisticLockingSetup(stageHarness, optimisticLockingTestSetup{
		Key1:                 watchedKey,
		Key2:                 unwatchedKey,
		KeyModifiedByClient2: watchedKey,
	}); err != nil {
		return err
	}

	stageHarness.Logger.Infof("Tearing down clients")

	watchedKey, unwatchedKey = allKeys[2], allKeys[3]
	return testOptimisticLockingSetup(stageHarness, optimisticLockingTestSetup{
		Key1:                 watchedKey,
		Key2:                 unwatchedKey,
		KeyModifiedByClient2: unwatchedKey,
	})
}

type optimisticLockingTestSetup struct {
	Key1                 string
	Key2                 string
	KeyModifiedByClient2 string
}

// testOptimisticLockingSetup runs a WATCH/MULTI/EXEC scenario with two clients.
// Client 1 sets Key1 and Key2, watches Key1, then queues a transaction that modifies Key2.
// Client 2 modifies KeyModifiedByClient2 before EXEC is called.
// If KeyModifiedByClient2 is Key1, the transaction aborts and Key2 retains its pre-transaction value.
// If KeyModifiedByClient2 is Key2, the transaction succeeds and Key2 holds the value set in the transaction.
func testOptimisticLockingSetup(stageHarness *test_case_harness.TestCaseHarness, setup optimisticLockingTestSetup) error {
	logger := stageHarness.Logger

	clients, err := SpawnClients(2, "localhost:6379", stageHarness, logger)
	if err != nil {
		return err
	}
	for _, c := range clients {
		defer c.Close()
	}

	client1 := clients[0]
	client2 := clients[1]

	key1InitialValue := testerutils_random.RandomInt(1, 100)
	key2InitialValue := testerutils_random.RandomInt(1, 100)
	newValueSetByClient2 := testerutils_random.RandomInt(200, 400)
	newValueSetByClient1InTransaction := testerutils_random.RandomInt(500, 1000)

	// Client 1: Set initial values for both keys
	setKeyTestCase := test_cases.MultiCommandTestCase{
		CommandWithAssertions: []test_cases.CommandWithAssertion{
			{
				Command:   []string{"SET", setup.Key1, strconv.Itoa(key1InitialValue)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
			{
				Command:   []string{"SET", setup.Key2, strconv.Itoa(key2InitialValue)},
				Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
			},
		},
	}

	if err := setKeyTestCase.RunAll(client1, logger); err != nil {
		return err
	}

	// Client 1: Watch Key1
	watchTestCase := test_cases.WatchTestCase{Keys: []string{setup.Key1}}

	if err := watchTestCase.Run(client1, logger); err != nil {
		return err
	}

	// Transaction aborts if client 2 modifies the watched key (Key1)
	transactionShouldAbort := setup.KeyModifiedByClient2 == setup.Key1
	var expectedResponseArray []resp_assertions.RESPAssertion
	if !transactionShouldAbort {
		expectedResponseArray = []resp_assertions.RESPAssertion{
			resp_assertions.NewSimpleStringAssertion("OK"),
		}
	}

	// Client 1: Queue a transaction that updates Key2
	transactionTestCase := test_cases.TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", setup.Key2, strconv.Itoa(newValueSetByClient1InTransaction)},
		},
		ExpectedResponseArray: expectedResponseArray,
	}
	if err := transactionTestCase.RunWithoutExec(client1, logger); err != nil {
		return err
	}

	// Client 2: Modify its designated key
	keyModificationTestCase := test_cases.SendCommandTestCase{
		Command:   "SET",
		Args:      []string{setup.KeyModifiedByClient2, strconv.Itoa(newValueSetByClient2)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	if err := keyModificationTestCase.Run(client2, logger); err != nil {
		return err
	}

	// Client 1: EXEC aborts if client 2 touched Key1, succeeds otherwise
	if err := transactionTestCase.RunExec(client1, logger); err != nil {
		return err
	}

	// Determine the expected value of Key2 after EXEC
	var expectedValueOfKey2 int
	if transactionShouldAbort {
		expectedValueOfKey2 = key2InitialValue
	} else {
		expectedValueOfKey2 = newValueSetByClient1InTransaction
	}

	return (&test_cases.SendCommandTestCase{
		Command:   "GET",
		Args:      []string{setup.Key2},
		Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(expectedValueOfKey2)),
	}).Run(client1, logger)
}
