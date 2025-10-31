package test_cases

import (
	"crypto/sha256"
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type AclGetuserTestCase struct {
	Username                 string
	FlagsExpectedToBePresent []string
	FlagsExpectedToBeAbsent  []string
	ExpectedPasswords        []string
}

// RunForFlagsTemplateOnly is used to run the following assertions:
// 1. First element is "flags"
// 2. Second element is a RESP array
func (t *AclGetuserTestCase) RunForFlagsTemplateOnly(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	clientLogger := client.GetLogger()

	aclGetUserTestCase := SendCommandTestCase{
		Command: "ACL",
		Args:    []string{"GETUSER", t.Username},
		Assertion: resp_assertions.ArrayElementsAssertion{
			IndexAssertionSpecifications: []resp_assertions.ArrayIndexAssertionSpecification{
				{
					Index:     0,
					Assertion: resp_assertions.NewStringAssertion("flags"),
					PreAssertionHook: func() {
						clientLogger.Infof("Checking if the first element is \"flags\"")
					},
					AssertionSuccessHook: func() {
						clientLogger.Successf("✔ First element is \"flags\"")
					},
				},
				{
					Index:     1,
					Assertion: resp_assertions.DataTypeAssertion{ExpectedType: resp_value.ARRAY},
					PreAssertionHook: func() {
						clientLogger.Infof("Checking if the second element is an array")
					},
					AssertionSuccessHook: func() {
						clientLogger.Successf("✔ Second element is an array")
					},
				},
			},
		},
	}

	return aclGetUserTestCase.Run(client, logger)
}

func (t *AclGetuserTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	arrayElementsAssertion := resp_assertions.ArrayElementsAssertion{}
	clientLogger := client.GetLogger()

	if t.FlagsExpectedToBePresent != nil || t.FlagsExpectedToBeAbsent != nil {
		t.addAssertionForFlags(&arrayElementsAssertion, clientLogger)
	}

	if t.ExpectedPasswords != nil {
		t.addAssertionForPasswords(&arrayElementsAssertion, clientLogger)
	}

	aclGetUserTestCase := SendCommandTestCase{
		Command:   "ACL",
		Args:      []string{"GETUSER", t.Username},
		Assertion: arrayElementsAssertion,
	}

	return aclGetUserTestCase.Run(client, logger)
}

func (t *AclGetuserTestCase) addAssertionForFlags(assertion *resp_assertions.ArrayElementsAssertion, logger *logger.Logger) {
	// Assert for flags
	// Assertion for "flags" as the first element
	assertion.IndexAssertionSpecifications = append(assertion.IndexAssertionSpecifications,
		resp_assertions.ArrayIndexAssertionSpecification{
			Index:     0,
			Assertion: resp_assertions.NewStringAssertion("flags"),
			PreAssertionHook: func() {
				logger.Infof("Checking if the first element is \"flags\"")
			},
			AssertionSuccessHook: func() {
				logger.Successf("✔ First element is \"flags\"")
			},
		})

	// Assert the type of 2nd element to be array
	assertion.IndexAssertionSpecifications = append(assertion.IndexAssertionSpecifications,
		resp_assertions.ArrayIndexAssertionSpecification{
			Index:     1,
			Assertion: resp_assertions.DataTypeAssertion{ExpectedType: resp_value.ARRAY},
			PreAssertionHook: func() {
				logger.Infof("Checking if the second element is an array")
			},
			AssertionSuccessHook: func() {
				logger.Successf("✔ Second element is an array")
			},
		})

	multiAssertionForFlagsStatus := resp_assertions.MultiAssertion{}

	// Assert the presence of expected flags
	for _, flagExpectedToBePresent := range t.FlagsExpectedToBePresent {
		multiAssertionForFlagsStatus.AssertionSpecifications = append(multiAssertionForFlagsStatus.AssertionSpecifications,
			resp_assertions.AssertionSpecification{
				Assertion: resp_assertions.BulkStringPresentInArrayAssertion{
					ExpectedString: flagExpectedToBePresent,
				},
				PreAssertionHook: func() {
					logger.Infof("Checking if flag '%s' is present in the flags array", flagExpectedToBePresent)
				},
				AssertionSuccessHook: func() {
					logger.Successf("✔ Flag '%s' is present in the flags array", flagExpectedToBePresent)
				},
			})
	}

	// Assert the presence of unexpected flags
	for _, flagExpectedToBeAbsent := range t.FlagsExpectedToBeAbsent {
		multiAssertionForFlagsStatus.AssertionSpecifications = append(multiAssertionForFlagsStatus.AssertionSpecifications,
			resp_assertions.AssertionSpecification{
				Assertion: resp_assertions.BulkStringAbsentFromArrayAssertion{
					UnexpectedString: flagExpectedToBeAbsent,
				},
				PreAssertionHook: func() {
					logger.Infof("Checking if flag '%s' is absent from the flags array", flagExpectedToBeAbsent)
				},
				AssertionSuccessHook: func() {
					logger.Successf("✔ Flag '%s' is absent in the flags array", flagExpectedToBeAbsent)
				},
			})
	}

	assertion.IndexAssertionSpecifications = append(assertion.IndexAssertionSpecifications,
		resp_assertions.ArrayIndexAssertionSpecification{
			Index:     1,
			Assertion: multiAssertionForFlagsStatus,
		})
}

func (t *AclGetuserTestCase) addAssertionForPasswords(assertion *resp_assertions.ArrayElementsAssertion, logger *logger.Logger) {
	assertion.IndexAssertionSpecifications = append(assertion.IndexAssertionSpecifications,
		resp_assertions.ArrayIndexAssertionSpecification{
			Index:     2,
			Assertion: resp_assertions.NewStringAssertion("passwords"),
			PreAssertionHook: func() {
				logger.Infof("Checking if the third element of the array is \"passwords\"")
			},
			AssertionSuccessHook: func() {
				logger.Successf("✔ Third element is \"passwords\"")
			},
		})

	passwordHashes := []string{}

	for _, password := range t.ExpectedPasswords {
		passwordSha256Hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
		passwordHashes = append(passwordHashes, string(passwordSha256Hash[:]))

	}

	assertion.IndexAssertionSpecifications = append(assertion.IndexAssertionSpecifications,
		resp_assertions.ArrayIndexAssertionSpecification{
			Index:     3,
			Assertion: resp_assertions.NewOrderedStringArrayAssertion(passwordHashes),
			PreAssertionHook: func() {
				if len(passwordHashes) == 0 {
					logger.Infof("Checking passwords array to be empty")
				} else {
					logger.Infof("Checking expected password hashes to be present in the passwords array")
				}
			},
			AssertionSuccessHook: func() {
				if len(passwordHashes) == 0 {
					logger.Successf("✔ Passwords array is an empty array")
				} else {
					logger.Successf("✔ Expected password hashes are present in the passwords array")
				}

			},
		})
}
