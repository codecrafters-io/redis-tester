package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/stretchr/testify/assert"
)

func TestRDBFileCreator(t *testing.T) {
	RDBFileCreator, err := NewRDBFileCreator()
	if err != nil {
		t.Fatalf("CodeCrafters Tester Error: %s", err)
	}

	randomKeyAndValue := testerutils_random.RandomWords(2)
	randomKey, randomValue := randomKeyAndValue[0], randomKeyAndValue[1]

	if err := RDBFileCreator.Write([]KeyValuePair{{key: randomKey, value: randomValue}}); err != nil {
		t.Fatalf("CodeCrafters Tester Error: %s", err)
	}

	fh, _ := os.Open(filepath.Join(RDBFileCreator.Dir, RDBFileCreator.Filename))
	defer fh.Close()
	fmt.Println("File content:")
	data, err := io.ReadAll(fh)
	if err != nil {
		t.Fatalf("CodeCrafters Tester Error: %s", err)
	}

	versionData := string(data[:9])
	magicString := versionData[0:5]
	version := versionData[5:9]
	assert.Equal(t, "REDIS", magicString)
	assert.Equal(t, "0011", version)
	t.Logf("Version: %s", version)
}
