package resp_utils

import (
	"strconv"
	"strings"
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
