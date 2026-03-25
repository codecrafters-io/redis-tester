package filesystem_assertion

import (
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

	// Handle empty case first: If we expect no commands, it's enough to notify
	// that the append-only file is non-empty
	if res := a.assertEmptyAofFileCase(fileContents, quotedPath); res.Err != nil {
		return res
	}

	// Decode all the commands in the file
	// Notify if any error is encountered while decoding commands
	res, decodedCommands := a.decodeCommandsFromAppendOnlyFile(fileContents)

	if res.Err != nil {
		return res
	}

	if res := a.assertCommandsArrayLength(decodedCommands); res.Err != nil {
		return res
	}

	return a.assertCommandsPosition(decodedCommands)
}

func (a AofAppendOnlyFileAssertion) assertEmptyAofFileCase(fileContents []byte, quotedPath string) FilesystemAssertionResult {
	if len(a.ExpectedCommands) != 0 {
		return FilesystemAssertionResult{}
	}

	if len(fileContents) > 0 {
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Expected append-only file %s to be empty, is not empty", quotedPath),
		}
	}

	return FilesystemAssertionResult{
		Logs: []FilesystemAssertionLog{
			NewFilesystemAssertionLog(
				_SUCCESS,
				"✔ Found no commands in append-only file",
			),
		},
	}
}

func (a AofAppendOnlyFileAssertion) decodeCommandsFromAppendOnlyFile(fileContents []byte) (FilesystemAssertionResult, [][]string) {
	decodedCommands, err := resp_decoder.DecodeCommandsFromAppendOnlyFile(fileContents)

	if err == nil {
		return FilesystemAssertionResult{}, decodedCommands
	}

	infoLogs := []FilesystemAssertionLog{
		NewFilesystemAssertionLog(_INFO, "Reading commands from append-only file"),
	}

	var decodeLogs []FilesystemAssertionLog

	for _, foundCommand := range decodedCommands {
		decodeLogs = append(
			decodeLogs,
			NewFilesystemAssertionLog(_INFO, fmt.Sprintf("Decoded command: %q", foundCommand)),
		)
	}

	return FilesystemAssertionResult{
		Logs: append(infoLogs, decodeLogs...),
		Err:  err,
	}, decodedCommands
}

func (a AofAppendOnlyFileAssertion) assertCommandsArrayLength(decoded [][]string) FilesystemAssertionResult {
	if len(decoded) == len(a.ExpectedCommands) {
		return FilesystemAssertionResult{}
	}

	var logs []FilesystemAssertionLog

	logs = append(logs, NewFilesystemAssertionLog(_SUCCESS, "Expected commands"))

	for _, cmd := range a.ExpectedCommands {
		logs = append(logs, NewFilesystemAssertionLog(_SUCCESS, strings.Join(cmd, " ")))
	}

	logs = append(logs, NewFilesystemAssertionLog(_ERROR, "Found commands:"))

	for _, cmd := range decoded {
		logs = append(logs, NewFilesystemAssertionLog(_ERROR, strings.Join(cmd, " ")))
	}

	return FilesystemAssertionResult{
		Logs: logs,
		Err: fmt.Errorf(
			"Expected %d commands to be present in the append-only file, found %d",
			len(a.ExpectedCommands),
			len(decoded),
		),
	}
}

func (a AofAppendOnlyFileAssertion) assertCommandsPosition(decoded [][]string) FilesystemAssertionResult {
	var foundCommandLogs []FilesystemAssertionLog

	for i, foundCommand := range decoded {
		foundCommandStr := strings.Join(foundCommand, " ")
		expectedCommand := a.ExpectedCommands[i]
		expectedCommandStr := strings.Join(expectedCommand, " ")

		if expectedCommandStr != foundCommandStr {
			return FilesystemAssertionResult{
				Logs: foundCommandLogs,
				Err: fmt.Errorf(
					"Expected command #%d to be %q, got %q", i+1, expectedCommandStr, foundCommandStr,
				),
			}
		}

		foundCommandLogs = append(foundCommandLogs, NewFilesystemAssertionLog(
			_SUCCESS,
			fmt.Sprintf("✔ Found command: %q", foundCommand),
		))
	}

	return FilesystemAssertionResult{
		Logs: foundCommandLogs,
	}
}
