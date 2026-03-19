package filesystem_assertion

import (
	"fmt"
	"os"

	"al.essio.dev/pkg/shellescape"
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

	// TODO: This assertion will suppport multiple commands in later PRs
	// Keeping this here for defensive programming
	if len(a.ExpectedCommands) != 0 {
		panic("Codecrafters Internal Error - AofAppendOnlyFileAssertion does not support commands yet!")
	}

	if len(fileContents) > 0 {
		return FileSystemAssertionResult{
			Err: fmt.Errorf("Expected append-only file %s to be empty, is not empty", quotedPath),
		}
	}

	return FileSystemAssertionResult{
		SuccessLog: fmt.Sprintf("✔ Append-only file %s is empty", quotedPath),
	}
}
