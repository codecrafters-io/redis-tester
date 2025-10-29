package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

// TODO: Remove later: I had a hard time designing the interface for this test case and its assertion (AclGetuserResponseAssertion)
// After much thought, I decided to settle with the current interface

// I am still not quite satisfied with it though! Especially the following parts:
// - the ExpectXYZ() and the checks are repeated in the AclGetuserResponseAssertion as well
// - Delivering error messages in nested arrays

// Will try a new approach tomorrow. (I can think of kafka-style asserter for nested arrays)

// Need some help on designing a better interface for this test case and its assertion

type AclGetuserTestCase struct {
	username          string
	expectedFlags     []string
	unexpectedFlags   []string
	expectedPasswords []string
}

func NewAclGetUserTestCase(username string) *AclGetuserTestCase {
	return &AclGetuserTestCase{
		username: username,
	}
}

func (t *AclGetuserTestCase) ExpectFlags(flags []string) *AclGetuserTestCase {
	if flags == nil {
		panic("Codecrafters Internal Error - Cannot expect nil array of flags in AclGetuserTestCase.ExpectFlags")
	}

	t.expectedFlags = flags
	return t
}

func (t *AclGetuserTestCase) UnexpectFlags(flags []string) *AclGetuserTestCase {
	if flags == nil {
		panic("Codecrafters Internal Error - Cannot expect nil array of flags in AclGetuserTestCase.UnexpectFlags")
	}

	t.unexpectedFlags = flags
	return t
}

func (t *AclGetuserTestCase) ExpectPasswords(passwords []string) *AclGetuserTestCase {
	if passwords == nil {
		panic("Codecrafters Internal Error - Cannot expect nil array of passwords in AclGetUserTestCase.ExpectPasswords")
	}

	t.expectedPasswords = passwords
	return t
}

// RunForFlagsTemplateOnly is used to run the following assertions:
// 1. First element is "flags"
// 2. Second element is a RESP array
func (t *AclGetuserTestCase) RunForFlagsTemplateOnly(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	sendCommandTestCase := SendCommandTestCase{
		Command: "ACL",
		Args:    []string{"GETUSER", t.username},
		Assertion: resp_assertions.AclGetUserResponseTemplateAssertion{
			AssertForFlags: true,
		},
	}

	return sendCommandTestCase.Run(client, logger)
}

func (t *AclGetuserTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	getuserResponseAssertion := resp_assertions.NewAclGetUserResponseAssertion()

	if t.expectedFlags != nil {
		getuserResponseAssertion.ExpectFlags(t.expectedFlags)
	}

	if t.unexpectedFlags != nil {
		getuserResponseAssertion.UnexpectFlags(t.unexpectedFlags)
	}

	if t.expectedPasswords != nil {
		getuserResponseAssertion.ExpectPasswords(t.expectedPasswords)
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "ACL",
		Args:      []string{"GETUSER", t.username},
		Assertion: getuserResponseAssertion,
	}

	return sendCommandTestCase.Run(client, logger)
}
