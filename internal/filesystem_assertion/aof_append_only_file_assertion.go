package filesystem_assertion

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"al.essio.dev/pkg/shellescape"
	resp_decoder "github.com/codecrafters-io/redis-tester/internal/resp/decoder"
	"github.com/kballard/go-shellquote"
)

type AofAppendOnlyFileAssertion struct {
	AbsolutePath     string
	ExpectedCommands [][]string

	accumulatedLogs []FilesystemAssertionLog
}

func (a *AofAppendOnlyFileAssertion) Run() FilesystemAssertionResult {
	if a.AbsolutePath == "" {
		panic("Codecrafters Internal Error - AofAppendOnlyFileAssertion: AbsolutePath cannot be empty")
	}

	// Reset the accumulated logs
	a.accumulatedLogs = []FilesystemAssertionLog{}

	quotedPath := shellescape.Quote(a.AbsolutePath)
	fileContents, err := os.ReadFile(a.AbsolutePath)

	if err != nil {
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Error reading append-only file %s: %w", quotedPath, err),
		}
	}

	// Assert the empty case first separatly, we need not decode commands here
	// Checking file contents is enough
	if err, done := a.assertEmptyAofFileCase(fileContents); done {
		return FilesystemAssertionResult{
			Logs: a.accumulatedLogs,
			Err:  err,
		}
	}

	decodedCommands, err := a.decodeCommandsFromAppendOnlyFile(fileContents)

	if err != nil {
		return FilesystemAssertionResult{
			Logs: a.accumulatedLogs,
			Err:  err,
		}
	}

	if err := a.assertCommandsArrayLength(decodedCommands); err != nil {
		return FilesystemAssertionResult{
			Logs: a.accumulatedLogs,
			Err:  err,
		}
	}

	if err := a.assertCommandsPosition(decodedCommands); err != nil {
		return FilesystemAssertionResult{
			Logs: a.accumulatedLogs,
			Err:  err,
		}
	}

	return FilesystemAssertionResult{
		Logs: a.accumulatedLogs,
		Err:  nil,
	}
}

func (a *AofAppendOnlyFileAssertion) assertEmptyAofFileCase(fileContents []byte) (error, bool) {
	if len(a.ExpectedCommands) != 0 {
		return nil, false
	}

	if len(fileContents) > 0 {
		return errors.New("Expected append-only file to be empty, is not empty"), true
	}

	a.registerLog(NewFilesystemAssertionLog(_SUCCESS, "✔ Found no commands in append-only file"))
	return nil, true
}

func (a *AofAppendOnlyFileAssertion) decodeCommandsFromAppendOnlyFile(fileContents []byte) ([][]string, error) {
	decodedCommands, err := resp_decoder.DecodeCommandsFromAppendOnlyFile(fileContents)

	if err == nil {
		return decodedCommands, nil
	}

	a.registerLog(NewFilesystemAssertionLog(_INFO, "Reading commands from append-only file"))

	for _, foundCommand := range decodedCommands {
		a.registerLog(
			NewFilesystemAssertionLog(_INFO, fmt.Sprintf("Decoded command: %q", foundCommand)),
		)
	}

	return decodedCommands, err
}

func (a *AofAppendOnlyFileAssertion) assertCommandsArrayLength(decoded [][]string) error {
	if len(decoded) == len(a.ExpectedCommands) {
		return nil
	}

	a.registerLog(NewFilesystemAssertionLog(_SUCCESS, "Expected commands"))

	for _, cmd := range a.ExpectedCommands {
		a.registerLog(NewFilesystemAssertionLog(_SUCCESS, strings.Join(cmd, " ")))
	}

	a.registerLog(NewFilesystemAssertionLog(_ERROR, "Found commands:"))

	for _, cmd := range decoded {
		a.registerLog(NewFilesystemAssertionLog(_ERROR, strings.Join(cmd, " ")))
	}

	return fmt.Errorf(
		"Expected %d commands to be present in the append-only file, found %d",
		len(a.ExpectedCommands),
		len(decoded),
	)
}

func (a *AofAppendOnlyFileAssertion) assertCommandsPosition(decoded [][]string) error {

	for i, foundCommand := range decoded {
		expectedCommand := a.ExpectedCommands[i]

		expectedCommandStr := shellquote.Join(expectedCommand...)
		foundCommandStr := shellquote.Join(foundCommand...)

		if slices.Equal(foundCommand, expectedCommand) {
			return fmt.Errorf(
				"Expected command #%d to be %q, got %q",
				i+1,
				expectedCommandStr,
				foundCommandStr,
			)
		}

		a.registerLog(NewFilesystemAssertionLog(
			_SUCCESS,
			fmt.Sprintf("✔ Found command: %q", foundCommand),
		))
	}

	return nil
}

func (a *AofAppendOnlyFileAssertion) registerLog(log FilesystemAssertionLog) {
	a.accumulatedLogs = append(a.accumulatedLogs, log)
}
