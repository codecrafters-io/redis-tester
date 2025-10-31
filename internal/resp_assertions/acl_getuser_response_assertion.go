package resp_assertions

import (
	"crypto/sha256"
	"fmt"
	"slices"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type passwordHashPair struct {
	Password string
	Hash     string
}

type AclGetuserResponseAssertion struct {
	expectedFlags     []string
	unexpectedFlags   []string
	expectedPasswords []passwordHashPair
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

	a.expectedPasswords = make([]passwordHashPair, 0)

	for _, password := range passwords {
		sha256Hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
		a.expectedPasswords = append(a.expectedPasswords, passwordHashPair{
			Password: password,
			Hash:     sha256Hash,
		})
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
	for _, expectedFlag := range a.expectedFlags {
		foundExpectedFlag := false

		for _, actualFlag := range flagsArray.Array() {
			if actualFlag.String() == expectedFlag {
				foundExpectedFlag = true
				break
			}
		}

		if !foundExpectedFlag {
			return fmt.Errorf("Flags array: Expected flag '%s' to be present in the array", expectedFlag)
		}

	}

	// Assert 'must be absent' flags
	for _, unexpectedFlag := range a.unexpectedFlags {
		for _, actualFlag := range flagsArray.Array() {
			if actualFlag.String() == unexpectedFlag {
				return fmt.Errorf("Flags array: Expected '%s' to be absent from the array", unexpectedFlag)
			}
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
	passwordHashes := []string{}

	for _, passwordHashPair := range a.expectedPasswords {
		passwordHashes = append(passwordHashes, passwordHashPair.Hash)
	}

	passwordsAssertion := NewOrderedStringArrayAssertion(passwordHashes)

	if err := passwordsAssertion.Run(passwordsArray); err != nil {
		return fmt.Errorf("Passwords array: %w", err)
	}

	return nil
}
