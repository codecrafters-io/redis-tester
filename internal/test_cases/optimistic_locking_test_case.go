package test_cases

import (
	"slices"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/random"
)

// OptimisticLockingTestCase tests Redis optimistic locking behavior using WATCH/MULTI/EXEC.
//
// It sets up a two-client scenario where:
//   - WatcherClient sets initial values for InitialKeys, watches KeysWatchedByWatcherClient,
//     and queues a transaction that modifies KeyToBeModifiedByWatcherClientInTransaction.
//   - ModifierClient modifies KeyToBeModifiedByModifierClient before EXEC is called.
//
// If KeyToBeModifiedByModifierClient is one of the watched keys, the transaction is
// expected to abort and EXEC returns a null array. Otherwise the transaction succeeds
// and EXEC returns an array of responses.
//
// After EXEC, the test verifies the value of KeyToBeModifiedByWatcherClientInTransaction
// reflects either the transaction's write (success) or its pre-transaction value (failure).
//
// Calling Run executes the full sequence. Individual Run* methods can be called separately
// to interleave custom steps such as UNWATCH or DISCARD between the standard steps.
type OptimisticLockingTestCase struct {
	WatcherClient                               *instrumented_resp_connection.InstrumentedRespConnection
	ModifierClient                              *instrumented_resp_connection.InstrumentedRespConnection
	InitialKeys                                 []string
	KeysWatchedByWatcherClient                  []string
	KeyToBeModifiedByWatcherClientInTransaction string
	KeyToBeModifiedByModifierClient             string

	// Used for tracking internal state
	initialValues            []int
	valueSetInTransaction    int
	valueSetByModifierClient int
	transactionTestCase      *TransactionTestCase
}

func (t *OptimisticLockingTestCase) Run(logger *logger.Logger) error {
	if err := t.RunSetInitialKeys(logger); err != nil {
		return err
	}

	if err := t.RunWatchKeys(logger); err != nil {
		return err
	}

	if err := t.RunTransactionWithoutExec(logger); err != nil {
		return err
	}

	if err := t.RunSetKeyUsingModifierClient(logger); err != nil {
		return err
	}

	if err := t.RunExec(logger); err != nil {
		return err
	}

	return t.RunValueCheckOfKeyModifiedInTransaction(logger)
}

// RunSetInitialKeys will set the initial values of keys using the modifier client
func (t *OptimisticLockingTestCase) RunSetInitialKeys(logger *logger.Logger) error {
	initialValues := random.RandomInts(1, 50, len(t.InitialKeys))

	var commandWithAssertions []CommandWithAssertion

	for i, initialKey := range t.InitialKeys {
		commandWithAssertions = append(commandWithAssertions, CommandWithAssertion{
			Command:   []string{"SET", initialKey, strconv.Itoa(initialValues[i])},
			Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
		})
	}

	setKeysTestCase := MultiCommandTestCase{
		CommandWithAssertions: commandWithAssertions,
	}

	t.initialValues = initialValues

	return setKeysTestCase.RunAll(t.WatcherClient, logger)
}

// RunWatchKeys will watch the specified keys using the watcher client
func (t *OptimisticLockingTestCase) RunWatchKeys(logger *logger.Logger) error {
	watchTestCase := WatchTestCase{
		Keys: t.KeysWatchedByWatcherClient,
	}

	return watchTestCase.Run(t.WatcherClient, logger)
}

// RunTransactionWithoutExec will run transaction that will modify the specified key using
// the watcher client without exec
func (t *OptimisticLockingTestCase) RunTransactionWithoutExec(logger *logger.Logger) error {
	valueUsedInTransaction := random.RandomInt(50, 100)
	var responsesArray []resp_assertions.RESPAssertion

	if !t.doesTransactionFail() {
		responsesArray = []resp_assertions.RESPAssertion{resp_assertions.NewSimpleStringAssertion("OK")}
	}

	transactionTestCase := TransactionTestCase{
		CommandQueue: [][]string{
			{"SET", t.KeyToBeModifiedByWatcherClientInTransaction, strconv.Itoa(valueUsedInTransaction)},
		},
		ExpectedResponseArray: responsesArray,
	}

	t.transactionTestCase = &transactionTestCase
	t.valueSetInTransaction = valueUsedInTransaction

	return transactionTestCase.RunWithoutExec(t.WatcherClient, logger)
}

// RunSetKeyUsingModifierClient will modify the specified key using the modifier client
func (t *OptimisticLockingTestCase) RunSetKeyUsingModifierClient(logger *logger.Logger) error {
	isWatchedKeyBeingModified := slices.Contains(t.KeysWatchedByWatcherClient, t.KeyToBeModifiedByModifierClient)

	if isWatchedKeyBeingModified {
		logger.Infof("Modifying the value of %s (watched key)", t.KeyToBeModifiedByModifierClient)
	} else {
		logger.Infof("Modifying the value of %s (unwatched key)", t.KeyToBeModifiedByModifierClient)
	}

	valueSetByModifierClient := random.RandomInt(100, 200)

	modifyValueTestCase := SendCommandTestCase{
		Command:   "SET",
		Args:      []string{t.KeyToBeModifiedByModifierClient, strconv.Itoa(valueSetByModifierClient)},
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	t.valueSetByModifierClient = valueSetByModifierClient

	return modifyValueTestCase.Run(t.ModifierClient, logger)
}

func (t *OptimisticLockingTestCase) RunExec(logger *logger.Logger) error {
	return t.transactionTestCase.RunExec(t.WatcherClient, logger)
}

func (t *OptimisticLockingTestCase) RunValueCheckOfKeyModifiedInTransaction(logger *logger.Logger) error {
	if t.doesTransactionFail() {
		logger.Infof("Checking if the transaction failed")
	} else {
		logger.Infof("Checking if the transaction succeeded")
	}

	var expectedValue int

	if !t.doesTransactionFail() {
		expectedValue = t.valueSetInTransaction
	} else if t.KeyToBeModifiedByModifierClient == t.KeyToBeModifiedByWatcherClientInTransaction {
		expectedValue = t.valueSetByModifierClient
	} else {
		expectedValueIdx := slices.Index(t.InitialKeys, t.KeyToBeModifiedByWatcherClientInTransaction)

		if expectedValueIdx == -1 {
			panic("Codecrafters Internal Error - t.KeyToBeModifiedByWatcherClientInTransaction not present inside t.InitialKeys")
		}

		expectedValue = t.initialValues[expectedValueIdx]
	}

	return (&SendCommandTestCase{
		Command:   "GET",
		Args:      []string{t.KeyToBeModifiedByWatcherClientInTransaction},
		Assertion: resp_assertions.NewBulkStringAssertion(strconv.Itoa(expectedValue)),
	}).Run(t.WatcherClient, logger)
}

func (t *OptimisticLockingTestCase) doesTransactionFail() bool {
	// Transaction fails if modifier client modifies any of the watched keys
	return slices.Contains(t.KeysWatchedByWatcherClient, t.KeyToBeModifiedByModifierClient)
}
