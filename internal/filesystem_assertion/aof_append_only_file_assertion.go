package filesystem_assertion

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"al.essio.dev/pkg/shellescape"
	resp_decoder "github.com/codecrafters-io/redis-tester/internal/resp/decoder"
	"github.com/dustin/go-humanize/english"
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

	a.registerLog(NewFilesystemAssertionLog(
		_SUCCESS,
		fmt.Sprintf("✔ Append-only file %s exists", a.AbsolutePath),
	))

	// Assert the empty case first separatly, we need not decode commands here
	// Checking file contents is enough
	if err, done := a.assertEmptyAofFileCase(fileContents); done {
		return FilesystemAssertionResult{
			Logs: a.accumulatedLogs,
			Err:  err,
		}
	}

	a.registerLog(NewFilesystemAssertionLog(_INFO, fmt.Sprintf("Reading commands from append-only file %s", quotedPath)))

	decodedCommands, err := a.decodeCommandsFromAppendOnlyFile(fileContents)

	if err != nil {
		return FilesystemAssertionResult{
			Logs: a.accumulatedLogs,
			Err:  err,
		}
	}

	decodedCommands = a.removeSelectCommandFromDecodedCommands(decodedCommands)

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

	a.registerLog(NewFilesystemAssertionLog(_SUCCESS, "✔ Append-only file is empty"))
	return nil, true
}

func (a *AofAppendOnlyFileAssertion) decodeCommandsFromAppendOnlyFile(fileContents []byte) ([][]string, error) {
	decodedCommands, err := resp_decoder.DecodeCommandsFromAppendOnlyFile(fileContents)

	if err == nil {
		return decodedCommands, nil
	}

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

	a.registerLog(NewFilesystemAssertionLog(_SUCCESS, "Expected commands:"))

	for i, cmd := range a.ExpectedCommands {
		a.registerLog(NewFilesystemAssertionLog(_SUCCESS, fmt.Sprintf("%d. %s", i+1, strings.Join(cmd, " "))))
	}

	a.registerLog(NewFilesystemAssertionLog(_ERROR, "Found commands:"))

	for i, cmd := range decoded {
		a.registerLog(NewFilesystemAssertionLog(_ERROR, fmt.Sprintf("%d. %s", i+1, strings.Join(cmd, " "))))
	}

	return fmt.Errorf(
		"Expected %s to be present in the append-only file, found %d",
		english.Plural(len(a.ExpectedCommands), "command", "commands"),
		len(decoded),
	)
}

func (a *AofAppendOnlyFileAssertion) assertCommandsPosition(decoded [][]string) error {

	for i, foundCommand := range decoded {
		expectedCommand := a.ExpectedCommands[i]

		expectedCommandStr := shellquote.Join(expectedCommand...)
		foundCommandStr := shellquote.Join(foundCommand...)

		if !slices.Equal(foundCommand, expectedCommand) {
			return fmt.Errorf(
				"Expected command #%d to be %q, got %q",
				i+1,
				expectedCommandStr,
				foundCommandStr,
			)
		}

		a.registerLog(NewFilesystemAssertionLog(
			_SUCCESS,
			fmt.Sprintf("✔ Found command: %q", foundCommandStr),
		))
	}

	return nil
}

func (a *AofAppendOnlyFileAssertion) registerLog(log FilesystemAssertionLog) {
	a.accumulatedLogs = append(a.accumulatedLogs, log)
}

func (a *AofAppendOnlyFileAssertion) removeSelectCommandFromDecodedCommands(decodedCommands [][]string) [][]string {
	if len(decodedCommands) != 0 && slices.Equal(decodedCommands[0], []string{"SELECT", "0"}) {
		decodedCommands = decodedCommands[1:]
	}
	return decodedCommands
}
