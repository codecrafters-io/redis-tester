package filesystem_assertion

import (
	"fmt"
	"os"

	"al.essio.dev/pkg/shellescape"
)

type DirExistsAssertion struct {
	AbsolutePath string
}

func (a *DirExistsAssertion) Run() FilesystemAssertionResult {
	if a.AbsolutePath == "" {
		panic("Codecrafters Internal Error - AbsolutePath in DirExistsAssertion cannot be empty")
	}

	info, err := os.Stat(a.AbsolutePath)
	escapedPath := shellescape.Quote(a.AbsolutePath)

	if err == nil {
		if !info.IsDir() {
			return FilesystemAssertionResult{
				Err: fmt.Errorf("Expected %s exists, but is not a directory", escapedPath),
			}
		}

		return FilesystemAssertionResult{
			Logs: []FilesystemAssertionLog{
				NewFilesystemAssertionLog(_SUCCESS, fmt.Sprintf("✔ Directory %s exists", escapedPath)),
			},
		}
	}

	// Does not exist
	if os.IsNotExist(err) {
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Expected directory %s to exist, not found", escapedPath),
		}
	}

	// Other errors
	return FilesystemAssertionResult{
		Err: fmt.Errorf("Error retrieving directory info of %s: %w", escapedPath, err),
	}
}
