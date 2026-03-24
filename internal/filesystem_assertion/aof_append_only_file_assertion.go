package filesystem_assertion

import (
	"fmt"
	"os"

	"al.essio.dev/pkg/shellescape"
	resp_decoder "github.com/codecrafters-io/redis-tester/internal/resp/decoder"
)

type AofAppendOnlyFileAssertion struct {
	AbsolutePath     string
	ExpectedCommands []string
}

func (a AofAppendOnlyFileAssertion) Run() FileSystemAssertionResult {
	if a.AbsolutePath == "" {
		panic("Codecrafters Internal Error - AofAppendOnlyFileAssertion: AbsolutePath cannot be empty")
	}

	quotedPath := shellescape.Quote(a.AbsolutePath)

	fileContents, err := os.ReadFile(a.AbsolutePath)

	if err != nil {
		return FileSystemAssertionResult{
			Err: fmt.Errorf("Error reading append-only file %s: %w", quotedPath, err),
		}
	}

	foundCommands, err := resp_decoder.DecodeCommandsFromAppendOnlyFile(fileContents)

	if err != nil {
		// Construct info logs from all the found commands
		var infoLogs []FileSystemAssertionLog

		for _, foundCommand := range foundCommands {
			infoLogs = append(
				infoLogs,
				NewFileSystemAssertionResultLog(
					_INFO,
					fmt.Sprintf("Found command: %q", foundCommand),
				),
			)
		}

		return FileSystemAssertionResult{
			Logs: infoLogs,
			Err:  err,
		}
	}

	if len(foundCommands) != len(a.ExpectedCommands) {
		return FileSystemAssertionResult{
			Err: fmt.Errorf(
				"Expected %d commands to be present in the append-only file, found %d",
				len(a.ExpectedCommands),
				len(foundCommands),
			),
		}
	}

	var successLogs []FileSystemAssertionLog

	for i, foundCommand := range foundCommands {
		expectedCommand := a.ExpectedCommands[i]

		if expectedCommand != foundCommand {
			return FileSystemAssertionResult{
				Logs: successLogs,
				Err: fmt.Errorf(
					"Expected command #%d to be %q, got %q", i+1, expectedCommand, foundCommand,
				),
			}
		} else {
			successLogs = append(successLogs, NewFileSystemAssertionResultLog(
				_SUCCESS,
				fmt.Sprintf("✔ Found command: %q", foundCommand),
			))
		}
	}

	return FileSystemAssertionResult{
		Logs: successLogs,
	}
}
