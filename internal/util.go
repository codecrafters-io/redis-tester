package internal

import (
	"os"
)

func deleteRDBfile() {
	fileName := "dump.rdb"
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(fileName)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
