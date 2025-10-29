package resp_assertions

import (
	"crypto/sha256"
	"fmt"
	"slices"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type AclGetuserResponseAssertion struct {
	expectedFlags     []string
	unexpectedFlags   []string
	expectedPasswords map[string]string
}

func NewAclGetUserResponseAssertion() *AclGetuserResponseAssertion {
	return &AclGetuserResponseAssertion{
		expectedFlags:     nil,
		expectedPasswords: nil,
	}
}

func (a *AclGetuserResponseAssertion) ExpectFlags(flags []string) *AclGetuserResponseAssertion {
	if flags == nil {
		panic("Codecrafters Internal Error - Cannot expect nil flags in AclGetuserResponseAssertion.ExpectFlags")
	}

	a.expectedFlags = flags
	return a
}

func (a *AclGetuserResponseAssertion) UnexpectFlags(flags []string) *AclGetuserResponseAssertion {
	if flags == nil {
		panic("Codecrafters Internal Error - flags is nil in AclGetuserResponseAssertion.UnexpectFlags")
	}

	for _, flag := range flags {
		if slices.Contains(a.expectedFlags, flag) {
			panic(fmt.Sprintf("Codecrafters Internal Error - Flag '%s' is present in expected flags and is being marked as unexpected", flag))
		}
	}

	a.unexpectedFlags = flags
	return a
}

func (a *AclGetuserResponseAssertion) ExpectPasswords(passwords []string) *AclGetuserResponseAssertion {
	if passwords == nil {
		panic("Codecrafters Internal Error - Cannot expect nil passwords in AclGetuserResponseAssertion.ExpectPasswords")
	}

	a.expectedPasswords = make(map[string]string)

	for _, password := range passwords {
		sha256Hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
		a.expectedPasswords[password] = sha256Hash
	}

	return a
}

func (a *AclGetuserResponseAssertion) Run(value resp_value.Value) error {
	if a.expectedFlags != nil || a.unexpectedFlags != nil {
		if err := a.assertFlags(value); err != nil {
			return err
		}
	}

	if a.expectedPasswords != nil {
		if err := a.assertPasswords(value); err != nil {
			return err
		}
	}

	return nil
}

func (a *AclGetuserResponseAssertion) assertFlags(value resp_value.Value) error {
	flagsTemplateAssertion := AclGetUserResponseTemplateAssertion{
		AssertForFlags: true,
	}

	if err := flagsTemplateAssertion.Run(value); err != nil {
		return err
	}

	flagsArray := value.Array()[1]

	// Assert 'must be present' flags
	for _, flag := range a.expectedFlags {
		flagPresentAssertion := NewStringPresentInArrayAssertion(flag)

		if err := flagPresentAssertion.Run(flagsArray); err != nil {
			return fmt.Errorf("Expected flag '%s' to be present in the flags array", flag)
		}
	}

	// Assert 'must be absent' flags
	for _, flag := range a.unexpectedFlags {
		flagAbsenseAssertion := NewStringAbsentInArrayAssertion(flag)

		if err := flagAbsenseAssertion.Run(flagsArray); err != nil {
			return fmt.Errorf("Expected flag '%s' to be absent in the flags array, but is present", flag)
		}
	}

	return nil
}

func (a *AclGetuserResponseAssertion) assertPasswords(value resp_value.Value) error {
	passwordsTemplateAssertion := AclGetUserResponseTemplateAssertion{
		AsserForPasswords: true,
	}

	if err := passwordsTemplateAssertion.Run(value); err != nil {
		return err
	}

	passwordsArray := value.Array()[3]

	if len(a.expectedPasswords) == 0 {
		emptyArrayAssertion := NewOrderedStringArrayAssertion([]string{})
		if err := emptyArrayAssertion.Run(passwordsArray); err != nil {
			return fmt.Errorf("Expected empty passwords array. Assertion failed with '%s'", err)
		}
		return nil
	}

	for password, passwordHash := range a.expectedPasswords {
		passwordHashPresentAssertion := NewStringPresentInArrayAssertion(passwordHash)
		if err := passwordHashPresentAssertion.Run(passwordsArray); err != nil {
			return fmt.Errorf("Expected hash of the password '%s' (%s) to be present in the passwords array", password, passwordHash)
		}
	}

	return nil
}
