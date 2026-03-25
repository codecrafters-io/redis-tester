package filesystem_assertion

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

type AofManifestFileAssertion struct {
	AbsolutePath           string
	AppendOnlyFileBasename string
}

func (a *AofManifestFileAssertion) Run() FilesystemAssertionResult {
	if a.AbsolutePath == "" || a.AppendOnlyFileBasename == "" {
		panic("Codecrafters Internal Error - AbsolutePath and AppendOnlyFileBaseName in AofManifestFileAssertion cannot be empty")
	}

	f, err := os.Open(a.AbsolutePath)
	quotedPath := shellescape.Quote(a.AbsolutePath)

	if err != nil {
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Error reading file %s: %s", quotedPath, err),
		}
	}

	defer f.Close()

	var foundLines []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		foundLines = append(foundLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return FilesystemAssertionResult{
			Err: fmt.Errorf("Error reading lines from the manifest file %s: %w", quotedPath, err),
		}
	}

	for _, foundLine := range foundLines {
		// Use strings.Contains because startOffset and endOffset is used by redis
		// to keep track of where the INCR file begins and ends in terms of replication
		if strings.Contains(foundLine, fmt.Sprintf("file %s seq 1 type i", a.AppendOnlyFileBasename)) {
			return FilesystemAssertionResult{
				Logs: []FilesystemAssertionLog{
					NewFilesystemAssertionLog(
						_SUCCESS,
						fmt.Sprintf("✔ Manifest file contains 'file %s seq 1 type i'", a.AppendOnlyFileBasename),
					),
				},
			}
		}
	}

	return FilesystemAssertionResult{
		Err: fmt.Errorf("Expected manifest file to contain 'file %s seq 1 type i', not found", a.AppendOnlyFileBasename),
	}
}
