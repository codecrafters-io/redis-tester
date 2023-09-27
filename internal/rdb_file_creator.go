package internal

import (
	"fmt"
	"os"
	"path/filepath"

	testerutils "github.com/codecrafters-io/tester-utils"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"

	"github.com/hdt3213/rdb/encoder"
)

type KeyValuePair struct {
	key      string
	value    string
	expiryTs int64 // Unix timestamp in milliseconds
}

type RDBFileCreator struct {
	Dir      string
	Filename string

	StageHarness *testerutils.StageHarness
}

func NewRDBFileCreator(stageHarness *testerutils.StageHarness) (*RDBFileCreator, error) {
	tmpDir, err := os.MkdirTemp("", "rdbfiles")
	if err != nil {
		return &RDBFileCreator{}, err
	}

	// On MacOS, the tmpDir is a symlink to a directory in /var/folders/...
	realDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return &RDBFileCreator{}, fmt.Errorf("could not resolve symlink: %v", err)
	}

	return &RDBFileCreator{
		Dir:      realDir,
		Filename: fmt.Sprintf("%s.rdb", testerutils_random.RandomWord()),
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

	if err := enc.WriteDBHeader(0, 5, 1); err != nil {
		return err
	}

	for _, keyValuePair := range keyValuePairs {
		if keyValuePair.expiryTs > 0 {
			if err := enc.WriteStringObject(keyValuePair.key, []byte(keyValuePair.value), encoder.WithTTL(uint64(keyValuePair.expiryTs))); err != nil {
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
