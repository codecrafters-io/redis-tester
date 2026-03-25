package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/filesystem_asserter"
	"github.com/codecrafters-io/redis-tester/internal/filesystem_assertion"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/tester-utils/logger"
)

type AofWriteTestCase struct {
	AppendOnlyFileAbsolutePath       string
	CommandWithAssertions            []CommandWithAssertion
	ExpectedCommandsInAppendOnlyFile [][]string
}

func (t *AofWriteTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	// Send all the commands first
	multiCommandTestCase := MultiCommandTestCase{
		CommandWithAssertions: t.CommandWithAssertions,
	}

	if err := multiCommandTestCase.RunAll(client, logger); err != nil {
		return err
	}

	// Assert which commands should appear in the appendonly file
	fsAsserter := filesystem_asserter.NewFilesystemAsserter([]filesystem_assertion.FilesystemAssertion{
		&filesystem_assertion.AofAppendOnlyFileAssertion{
			AbsolutePath:     t.AppendOnlyFileAbsolutePath,
			ExpectedCommands: t.ExpectedCommandsInAppendOnlyFile,
		},
	})

	return fsAsserter.RunAssertions(logger)
}
