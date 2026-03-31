package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	encoder "github.com/codecrafters-io/redis-tester/internal/resp/encoder"
	value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
	"github.com/dustin/go-humanize/english"
)

// AofDirectoryCreator is used to create an append-only directory
// Redis uses the same name for mannifest and the append-only file
// Eg. foo.manifest and foo.1.incr.aof
// But since this is used for testing user's code, to ensure that the users actually read from the
// manifest file and parse the append-only file name, AppendFilenameFlag and
// AppendOnlyFilenameInManifest can be specified separately
type AofDirectoryCreator struct {
	DataDirectory                string     // directory inside which Aof directory is created
	AppendDirName                string     // Value of appendonlydir flag
	AppendFileNameInFlag         string     // Value of appendfilename (Used for manifest file name)
	AppendOnlyFileNameInManifest string     // Value appendfilename to be used inside manifest
	CommandsInsideAppendOnlyFile [][]string // slice of commands to be written to append-only file
}

func (a *AofDirectoryCreator) Create(logger *logger.Logger) error {
	a.verifyMemberValues()

	appendDirPath := filepath.Join(a.DataDirectory, a.AppendDirName)
	manifestFileName := a.AppendFileNameInFlag + ".manifest"
	manifestFilePath := filepath.Join(appendDirPath, manifestFileName)
	actualAppendFileName := fmt.Sprintf("%s.1.incr.aof", a.AppendOnlyFileNameInManifest)
	actualAppendFilePath := filepath.Join(appendDirPath, actualAppendFileName)
	manifestFileEntry := fmt.Sprintf("file %s seq 1 type i", actualAppendFileName)

	if err := a.createAppendOnlyDirectory(logger, appendDirPath, manifestFileName, actualAppendFileName); err != nil {
		return err
	}

	if err := a.createAppendOnlyFile(logger, actualAppendFilePath); err != nil {
		return err
	}

	if err := a.createManifestFile(logger, manifestFilePath, manifestFileEntry); err != nil {
		return err
	}

	return nil
}

func (a *AofDirectoryCreator) createAppendOnlyDirectory(logger *logger.Logger, appendDirPath, manifestFileName, actualAppendFileName string) error {
	logger.Infof("Creating append-only directory %q:", a.AppendDirName)

	logger.WithAdditionalSecondaryPrefix(a.AppendDirName, func() {
		logger.Infof("  - %s", manifestFileName)
		logger.Infof("  - %s", actualAppendFileName)
	})

	if err := os.MkdirAll(appendDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create append-only directory %s: %w", appendDirPath, err)
	}

	return nil
}

func (a *AofDirectoryCreator) createAppendOnlyFile(logger *logger.Logger, actualAppendFilePath string) error {
	actualAppendFileName := filepath.Base(actualAppendFilePath)

	if len(a.CommandsInsideAppendOnlyFile) > 0 {
		logger.Infof(
			"Writing %s to append-only file %q",
			english.Plural(len(a.CommandsInsideAppendOnlyFile), "command", "commands"),
			actualAppendFileName,
		)
	} else {
		logger.Infof("Creating empty append-only file %s", actualAppendFileName)
	}

	var aofFileContents []byte

	for _, command := range a.CommandsInsideAppendOnlyFile {
		commandRespBytes := a.encodeCommandAsRESPBytes(command)
		aofFileContents = append(aofFileContents, commandRespBytes...)

		// Display the command as if it would be displayed using the quoted "%q" directive
		// But remove the surrounding quotes
		comandRespBytesFormatted := strings.Trim(
			fmt.Sprintf("%q", commandRespBytes),
			"\"",
		)

		logger.WithAdditionalSecondaryPrefix(actualAppendFileName, func() {
			logger.Infof("%s", comandRespBytesFormatted)
		})
	}

	if err := os.WriteFile(actualAppendFilePath, aofFileContents, 0o644); err != nil {
		return fmt.Errorf("Failed to create append-only file %s: %w", actualAppendFilePath, err)
	}

	return nil
}

func (a *AofDirectoryCreator) createManifestFile(logger *logger.Logger, manifestFilePath, manifestFileEntry string) error {
	manifestFileName := filepath.Base(manifestFilePath)

	logger.Infof("Creating manifest file %q", manifestFileName)

	logger.WithAdditionalSecondaryPrefix(manifestFileName, func() {
		logger.Infof("%s", manifestFileEntry)
	})

	manifestFileRawBytes := manifestFileEntry + "\n"
	if err := os.WriteFile(manifestFilePath, []byte(manifestFileRawBytes), 0o644); err != nil {
		return fmt.Errorf("Failed to create manifest file %s: %w", manifestFilePath, err)
	}
	return nil
}

func (a *AofDirectoryCreator) Cleanup(stageHarness *test_case_harness.TestCaseHarness) error {
	return os.RemoveAll(filepath.Join(a.DataDirectory, a.AppendDirName))
}

func (a *AofDirectoryCreator) verifyMemberValues() {
	if a.DataDirectory == "" {
		panic("Codecrafters Internal Error - DataDirectory cannot be empty in AofDirectoryCreator")
	}

	if a.AppendDirName == "" {
		panic("Codecrafters Internal Error - AppendDirName cannot be empty in AofDirectoryCreator")
	}

	if a.AppendFileNameInFlag == "" {
		panic("Codecrafters Internal Error - AppendFileName cannot be empty in AofDirectoryCreator")
	}

	if a.AppendOnlyFileNameInManifest == "" {
		panic("Codecrafters Internal Error - AppendOnlyFileNameInManifest cannot be empty in AofDirectoryCreator")
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

// encodeCommandAsRESPBytes encodes a given command as RESP bytes to be written to the append-only file
func (a *AofDirectoryCreator) encodeCommandAsRESPBytes(command []string) []byte {
	return encoder.Encode(value.NewStringArrayValue(command))
}
