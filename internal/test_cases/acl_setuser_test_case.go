package test_cases

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type AclSetuserTestCase struct {
	Username  string
	Passwords []string
}

func (t AclSetuserTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	passwordRules := []string{}

	for _, password := range t.Passwords {
		passwordRules = append(passwordRules, fmt.Sprintf(">%s", password))
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "ACL",
		Args:      append([]string{"SETUSER", t.Username}, passwordRules...),
		Assertion: resp_assertions.NewSimpleStringAssertion("OK"),
	}

	return sendCommandTestCase.Run(client, logger)
}
