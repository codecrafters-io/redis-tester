package resp_utils

import (
	"strconv"
	"strings"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

func GetByteOffset(args []string) int {
	offset := 0
	offset += 2 * (2*len(args) + 1)
	offset += (len(strconv.Itoa(len(args))) + 1)
	for _, arg := range args {
		msgLen := len(arg)
		offset += (len(strconv.Itoa(msgLen)) + 1)
		offset += (msgLen)
	}

	return offset
}

func GetByteOffsetHelper(args string) int {
	// Takes a string of the type "[ 'COMMAND', 'ARGS']"
	return GetByteOffset(strings.Split(args[1:len(args)-1], ", "))
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
