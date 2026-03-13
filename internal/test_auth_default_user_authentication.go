package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testdefaultUserAuthentication(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)

	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	firstClient, err := clientsSpawner.SpawnNextClient()
	if err != nil {
		return err
	}

	// Set default user password
	password := fmt.Sprintf("%s-%d", random.RandomWord(), random.RandomInt(1, 1000))

	aclSetUserTestCase := test_cases.AclSetuserTestCase{
		Username:  "default",
		Passwords: []string{password},
	}

	if err := aclSetUserTestCase.Run(firstClient, logger); err != nil {
		return err
	}

	// Run ACL WHOAMI from the existing client
	whoamiTestCase := test_cases.AclWhoamiTestCase{
		ExpectedUsername: "default",
	}

	if err := whoamiTestCase.Run(firstClient, logger); err != nil {
		return err
	}

	secondClient, err := clientsSpawner.SpawnNextClient()
	if err != nil {
		return err
	}

	whoamiNoauthTestCase := test_cases.AclWhoamiErrorTestCase{
		ExpectedErrorBeginsWith: "NOAUTH ",
	}

	return whoamiNoauthTestCase.Run(secondClient, logger)
}
