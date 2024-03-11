package internal

import (
	"os"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func deleteRDBfile() {
	fileName := "dump.rdb"
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(fileName)
}

func RandomAlphanumericString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		charIndex := testerutils_random.RandomInt(0, len(charset))
		result[i] = charset[charIndex]
	}
	return string(result)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
