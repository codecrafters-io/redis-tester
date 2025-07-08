package test_cases

import (
	"errors"
	"strconv"
	"strings"

	resp_connection "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GetAckTestCase struct{}

func (t GetAckTestCase) Run(client *resp_connection.RespConnection, logger *logger.Logger, offset int) error {
	commandTest := SendCommandTestCase{
		Command:                "REPLCONF",
		Args:                   []string{"GETACK", "*"},
		Assertion:              resp_assertions.NewCommandAssertion("REPLCONF", "ACK", strconv.Itoa(offset)),
		BetterErrorMessageFunc: betterGetAckErrorMessage,
	}

	return commandTest.Run(client, logger)
}

func betterGetAckErrorMessage(err error) error {
	if strings.HasPrefix(err.Error(), `Received: "" (no content received)`) {
		return errors.New(`
❌ Error: The master did not receive a response to "REPLCONF GETACK *"

💡 Hints:

• A single TCP read may contain multiple commands.
  There's no guarantee each read maps to exactly one command.

• It's possible the replica read the master’s handshake responses 
  with the REPLCONF command in one go.

• If the replica ignored the master’s handshake responses, 
  it may have inadvertently ignored the REPLCONF command as well.
`)
	}

	return err
}
