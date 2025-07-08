package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"

	"github.com/hdt3213/rdb/encoder"
)

type KeyValuePair struct {
	key      string
	value    string
	expiryTS int64 // Unix timestamp in milliseconds
}

type RDBFileCreator struct {
	Dir      string
	Filename string

	StageHarness *test_case_harness.TestCaseHarness
}

func NewRDBFileCreator() (*RDBFileCreator, error) {
	tmpDir, err := MkdirTemp("rdb")
	if err != nil {
		return &RDBFileCreator{}, err
	}

	return &RDBFileCreator{
		Dir:      tmpDir,
		Filename: fmt.Sprintf("%s.rdb", random.RandomWord()),
	}, nil
}

func (r *RDBFileCreator) Cleanup() {
	os.RemoveAll(r.Dir)
}

func (r *RDBFileCreator) Write(keyValuePairs []KeyValuePair) error {
	rdbFile, err := os.Create(filepath.Join(r.Dir, r.Filename))
	if err != nil {
		return err
	}
	defer rdbFile.Close()

	enc := encoder.NewEncoder(rdbFile)

	if err := enc.WriteHeader(); err != nil {
		return err
	}

	auxMap := map[string]string{
		"redis-ver":  "7.2.0",
		"redis-bits": "64",
	}

	for k, v := range auxMap {
		if err := enc.WriteAux(k, v); err != nil {
			return err
		}
	}

	keysWithTTLCount := 0

	for _, keyValuePair := range keyValuePairs {
		if keyValuePair.expiryTS > 0 {
			keysWithTTLCount++
		}
	}

	if err := enc.WriteDBHeader(0, uint64(len(keyValuePairs)), uint64(keysWithTTLCount)); err != nil {
		return err
	}

	for _, keyValuePair := range keyValuePairs {
		if keyValuePair.expiryTS > 0 {
			if err := enc.WriteStringObject(keyValuePair.key, []byte(keyValuePair.value), encoder.WithTTL(uint64(keyValuePair.expiryTS))); err != nil {
				return err
			}
		} else {
			if err := enc.WriteStringObject(keyValuePair.key, []byte(keyValuePair.value)); err != nil {
				return err
			}
		}
	}

	if err = enc.WriteEnd(); err != nil {
		return err
	}

	return nil
}

func (r *RDBFileCreator) Contents() ([]byte, error) {
	contents, err := os.ReadFile(filepath.Join(r.Dir, r.Filename))
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (r *RDBFileCreator) PrintContentHexdump(logger *logger.Logger) error {
	contents, err := r.Contents()
	if err != nil {
		return err
	}
	logger.Debugf("Hexdump of RDB file contents: \n%v\n", GetFormattedHexdump(contents))
	return nil
}
