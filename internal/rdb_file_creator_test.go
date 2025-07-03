package internal

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/codecrafters-io/tester-utils/random"
	"github.com/stretchr/testify/assert"
)

func TestRDBFileCreator(t *testing.T) {
	random.Init()

	RDBFileCreator, err := NewRDBFileCreator()
	if err != nil {
		t.Fatalf("CodeCrafters Tester Error: %s", err)
	}

	keyAndValue := random.RandomWords(2)
	key, value := keyAndValue[0], keyAndValue[1]

	if err := RDBFileCreator.Write([]KeyValuePair{{key: key, value: value}}); err != nil {
		t.Fatalf("CodeCrafters Tester Error: %s", err)
	}

	fh, _ := os.Open(filepath.Join(RDBFileCreator.Dir, RDBFileCreator.Filename))
	defer fh.Close()
	data, err := io.ReadAll(fh)
	if err != nil {
		t.Fatalf("CodeCrafters Tester Error: %s", err)
	}

	versionData := string(data[:9])
	magicString := versionData[0:5]
	version := versionData[5:9]
	assert.Equal(t, "REDIS", magicString)
	assert.Equal(t, "0011", version)
}
