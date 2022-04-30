package internal

import (
	testerutils "github.com/codecrafters-io/tester-utils"
	"strings"
)

func logFriendlyError(logger *testerutils.Logger, err error) {
	if err.Error() == "EOF" {
		logger.Infof("Hint: EOF is short for 'end of file'. This usually means that your program either:")
		logger.Infof(" (a) didn't send a complete response, or")
		logger.Infof(" (b) closed the connection early")
	}

	if strings.Contains(err.Error(), "connection reset by peer") {
		logger.Infof("Hint: 'connection reset by peer' usually means that your program closed the connection before sending a complete response.")
	}

	if strings.Contains(err.Error(), "reply is empty") {
		logger.Infof("Hint: 'reply is empty' usually means that your program sent an additional `\\n` in the response.")
		logger.Infof("       A common reason for this is using methods like `Println` that append a newline charater.")
	}
}
