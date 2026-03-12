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
		Logger:       logger,
	}
	firstClients, err := clientsSpawner.SpawnClients(1)
	if err != nil {
		return err
	}
	firstClient := firstClients[0]

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

	secondClients, err := clientsSpawner.SpawnClients(1)
	if err != nil {
		return err
	}
	secondClient := secondClients[0]

	whoamiNoauthTestCase := test_cases.AclWhoamiErrorTestCase{
		ExpectedErrorPattern: "^NOAUTH.*",
	}

	return whoamiNoauthTestCase.Run(secondClient, logger)
}
