package filesystem_assertion

import (
	"fmt"
	"os"

	"al.essio.dev/pkg/shellescape"
)

type DirExistsAssertion struct {
	AbsolutePath string
}

func (a DirExistsAssertion) Run() FileSystemAssertionResult {
	if a.AbsolutePath == "" {
		panic("Codecrafters Internal Error - AbsolutePath in DirExistsAssertion cannot be empty")
	}

	info, err := os.Stat(a.AbsolutePath)
	escapedPath := shellescape.Quote(a.AbsolutePath)

	if err == nil {
		return FileSystemAssertionResult{
			SuccessLog: fmt.Sprintf("✔ Directory %s exists", escapedPath),
		}
	}

	// Does not exist
	if os.IsNotExist(err) {
		return FileSystemAssertionResult{
			Err: fmt.Errorf("Expected directory %q does not exist", escapedPath),
		}
	}

	// Exists but is not a directory (Possible error)
	if !info.IsDir() {
		return FileSystemAssertionResult{
			Err: fmt.Errorf("Expected %s exists, but is not a directory", escapedPath),
		}
	}

	// Other errors
	return FileSystemAssertionResult{
		Err: fmt.Errorf("Error retrieving directory info of %s: %w", escapedPath, err),
	}
}
