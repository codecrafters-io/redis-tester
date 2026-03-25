package filesystem_assertion

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"al.essio.dev/pkg/shellescape"
	resp_decoder "github.com/codecrafters-io/redis-tester/internal/resp/decoder"
)

type AofAppendOnlyFileAssertion struct {
	AbsolutePath     string
	ExpectedCommands [][]string
}

func (a AofAppendOnlyFileAssertion) Run() FilesystemAssertionResult {
	if a.AbsolutePath == "" {
		panic("Codecrafters Internal Error - AofAppendOnlyFileAssertion: AbsolutePath cannot be empty")
	}

	quotedPath := shellescape.Quote(a.AbsolutePath)

	fileContents, err := os.ReadFile(a.AbsolutePath)

	if err != nil {
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Error reading append-only file %s: %w", quotedPath, err),
		}
	}

	// Assert the empty case first separatly, we need not decode commands here
	// Checking file contents is enough
	if result, done := a.assertEmptyAofFileCase(fileContents); done {
		return result
	}

	var assertionLogs []FilesystemAssertionLog
	// Decode all the commands in the file
	// Notify if any error is encountered while decoding commands
	res, decodedCommands := a.decodeCommandsFromAppendOnlyFile(fileContents, assertionLogs)

	if res.Err != nil {
		return res
	}

	if res := a.assertCommandsArrayLength(decodedCommands, assertionLogs); res.Err != nil {
		return res
	}

	return a.assertCommandsPosition(decodedCommands, assertionLogs)
}

func (a AofAppendOnlyFileAssertion) assertEmptyAofFileCase(fileContents []byte) (FilesystemAssertionResult, bool) {
	if len(a.ExpectedCommands) != 0 {
		return FilesystemAssertionResult{}, false
	}

	if len(fileContents) > 0 {
		return FilesystemAssertionResult{
			Err: errors.New("Expected append-only file to be empty, is not empty"),
		}, true
	}

	return FilesystemAssertionResult{
		Logs: []FilesystemAssertionLog{NewFilesystemAssertionLog(
			_SUCCESS,
			"✔ Found no commands in append-only file",
		),
		},
	}, true
}

func (a AofAppendOnlyFileAssertion) decodeCommandsFromAppendOnlyFile(fileContents []byte, previousLogs []FilesystemAssertionLog) (FilesystemAssertionResult, [][]string) {
	decodedCommands, err := resp_decoder.DecodeCommandsFromAppendOnlyFile(fileContents)

	if err == nil {
		return FilesystemAssertionResult{
			Logs: previousLogs,
		}, decodedCommands
	}

	allLogs := append(
		previousLogs,
		[]FilesystemAssertionLog{
			NewFilesystemAssertionLog(_INFO, "Reading commands from append-only file"),
		}...,
	)

	for _, foundCommand := range decodedCommands {
		allLogs = append(
			allLogs,
			NewFilesystemAssertionLog(_INFO, fmt.Sprintf("Decoded command: %q", foundCommand)),
		)
	}

	return FilesystemAssertionResult{
		Logs: allLogs,
		Err:  err,
	}, decodedCommands
}

func (a AofAppendOnlyFileAssertion) assertCommandsArrayLength(decoded [][]string, previousLogs []FilesystemAssertionLog) FilesystemAssertionResult {
	if len(decoded) == len(a.ExpectedCommands) {
		return FilesystemAssertionResult{}
	}

	allLogs := previousLogs
	allLogs = append(allLogs, NewFilesystemAssertionLog(_SUCCESS, "Expected commands"))

	for _, cmd := range a.ExpectedCommands {
		allLogs = append(allLogs, NewFilesystemAssertionLog(_SUCCESS, strings.Join(cmd, " ")))
	}

	allLogs = append(allLogs, NewFilesystemAssertionLog(_ERROR, "Found commands:"))

	for _, cmd := range decoded {
		allLogs = append(allLogs, NewFilesystemAssertionLog(_ERROR, strings.Join(cmd, " ")))
	}

	return FilesystemAssertionResult{
		Logs: allLogs,
		Err: fmt.Errorf(
			"Expected %d commands to be present in the append-only file, found %d",
			len(a.ExpectedCommands),
			len(decoded),
		),
	}
}

func (a AofAppendOnlyFileAssertion) assertCommandsPosition(decoded [][]string, previousLogs []FilesystemAssertionLog) FilesystemAssertionResult {
	allLogs := previousLogs

	for i, foundCommand := range decoded {
		foundCommandStr := strings.Join(foundCommand, " ")
		expectedCommand := a.ExpectedCommands[i]
		expectedCommandStr := strings.Join(expectedCommand, " ")

		if expectedCommandStr != foundCommandStr {
			return FilesystemAssertionResult{
				Logs: allLogs,
				Err: fmt.Errorf(
					"Expected command #%d to be %q, got %q", i+1, expectedCommandStr, foundCommandStr,
				),
			}
		}

		allLogs = append(allLogs, NewFilesystemAssertionLog(
			_SUCCESS,
			fmt.Sprintf("✔ Found command: %q", foundCommand),
		))
	}

	return FilesystemAssertionResult{
		Logs: allLogs,
	}
}
