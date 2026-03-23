package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	resp_encoder "github.com/codecrafters-io/redis-tester/internal/resp/encoder"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

// AofDirectoryCreator is used to create an append-only directory
// Redis uses the same name for mannifest and the append-only file
// Eg. foo.manifest and foo.1.incr.aof
// But since this is used for testing user's code, to ensure that the users actually read from the
// manifest file and parse the append-only file name, AppendFilenameFlag and
// AppendOnlyFilenameInManifest can be specified separately
type AofDirectoryCreator struct {
	WorkingDirectory             string     // directory inside which Aof directory is created
	AppendDirName                string     // Value of appendonlydir flag
	AppendFileNameinFlag         string     // Value of appendfilename (Used for manifest file name)
	AppendOnlyFilenameInManifest string     // Value appendfilename to be used inside manifest
	CommandsInsideAppendOnlyFile [][]string // slice of commands to be written to append-only file
}

func (a *AofDirectoryCreator) Create(logger *logger.Logger) error {
	a.verifyMemberValues()

	appendDirPath := filepath.Join(a.WorkingDirectory, a.AppendDirName)
	manifestFileName := a.AppendFileNameinFlag + ".manifest"
	manifestFilePath := filepath.Join(appendDirPath, manifestFileName)
	actualAppendFileName := fmt.Sprintf("%s.1.incr.aof", a.AppendOnlyFilenameInManifest)
	actualAppendFilePath := filepath.Join(appendDirPath, actualAppendFileName)
	manifestFileEntry := fmt.Sprintf("file %s seq 1 type i", actualAppendFileName)

	if err := os.MkdirAll(appendDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create append-only directory %s: %w", appendDirPath, err)
	}

	appendBody := a.EncodeCommandsAsRESP(a.CommandsInsideAppendOnlyFile)
	if err := os.WriteFile(actualAppendFilePath, appendBody, 0o644); err != nil {
		return fmt.Errorf("Failed to create append-only file %s: %w", actualAppendFilePath, err)
	}

	manifestRaw := manifestFileEntry + "\n"
	if err := os.WriteFile(manifestFilePath, []byte(manifestRaw), 0o644); err != nil {
		return fmt.Errorf("Failed to create manifest file %s: %w", manifestFilePath, err)
	}

	logger.Infof("Creating append-only directory %s:", a.AppendDirName)
	logger.WithAdditionalSecondaryPrefix(a.AppendDirName, func() {
		logger.Infof("  - %s", manifestFileName)
		logger.Infof("  - %s", actualAppendFileName)
	})

	logger.Infof("Creating manifest file %s", manifestFileName)
	logger.WithAdditionalSecondaryPrefix(manifestFileName, func() {
		logger.Infof("%s", manifestFileEntry)
	})

	logger.Infof("Writing the following commands to append-only file %s", actualAppendFileName)
	logger.WithAdditionalSecondaryPrefix(actualAppendFileName, func() {
		for _, cmd := range a.CommandsInsideAppendOnlyFile {
			logger.Infof("%s", strings.Join(cmd, " "))
		}
	})

	return nil
}

func (a *AofDirectoryCreator) Cleanup(stageHarness *test_case_harness.TestCaseHarness) error {
	return os.RemoveAll(filepath.Join(a.WorkingDirectory, a.AppendDirName))
}

func (a *AofDirectoryCreator) verifyMemberValues() {
	if a.WorkingDirectory == "" {
		panic("Codecrafters Internal Error - WorkingDirectory cannot be empty in AofDirectoryCreator")
	}

	if a.AppendDirName == "" {
		panic("Codecrafters Internal Error - AppendDirName cannot be empty in AofDirectoryCreator")
	}

	if a.AppendFileNameinFlag == "" {
		panic("Codecrafters Internal Error - AppendFileName cannot be empty in AofDirectoryCreator")
	}

	if a.AppendOnlyFilenameInManifest == "" {
		panic("Codecrafters Internal Error - AppendOnlyFileNameInManifest cannot be empty in AofDirectoryCreator")
	}

	if len(a.CommandsInsideAppendOnlyFile) == 0 {
		panic("Codecrafters Internal Error - CommandsInsideAppendOnlyFile cannot be nil or empty in AofDirectoryCreator")
	}

	for i, cmd := range a.CommandsInsideAppendOnlyFile {
		if len(cmd) == 0 {
			panic(
				fmt.Sprintf(
					"Codecrafters Internal Error - CommandsInsideAppendOnlyFile[%d] is empty in AofDirectoryCreator",
					i,
				),
			)
		}
	}
}

// EncodeCommandsAsRESP encodes commands as RESP bytes to be written to the append-only file
func (a *AofDirectoryCreator) EncodeCommandsAsRESP(commands [][]string) []byte {
	var out []byte

	for _, cmd := range commands {
		out = append(out, resp_encoder.Encode(resp_value.NewStringArrayValue(cmd))...)
	}

	return out
}
