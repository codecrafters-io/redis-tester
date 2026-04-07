package filesystem_assertion

import (
	"fmt"
	"os"

	"al.essio.dev/pkg/shellescape"
)

type DirDoesNotExistAssertion struct {
	AbsolutePath string
}

func (a *DirDoesNotExistAssertion) Run() FilesystemAssertionResult {
	if a.AbsolutePath == "" {
		panic("Codecrafters Internal Error - AbsolutePath in DirDoesNotExistAssertion cannot be empty")
	}

	info, err := os.Stat(a.AbsolutePath)
	escapedPath := shellescape.Quote(a.AbsolutePath)

	if err != nil {
		if os.IsNotExist(err) {
			return FilesystemAssertionResult{
				Logs: []FilesystemAssertionLog{
					NewFilesystemAssertionLog(_SUCCESS, fmt.Sprintf("✔ Directory %s does not exist", escapedPath)),
				},
			}
		}
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Error retrieving directory info of %s: %w", escapedPath, err),
		}
	}

	if info.IsDir() {
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Expected directory %s to not exist, found directory", escapedPath),
		}
	}

	return FilesystemAssertionResult{
		Err: fmt.Errorf("Expected directory %s to not exist, found non-directory file", escapedPath),
	}
}
