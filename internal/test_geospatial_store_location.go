package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialStoreLocation(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	locationKey := testerutils_random.RandomWord()
	location := data_structures.GenerateRandomLocations(1)[0]

	// Add a location
	geoAddTestCase := test_cases.NewGeoAddTestCaseWithValidCoordinates(locationKey, location, 1)
	if err := geoAddTestCase.Run(client, logger); err != nil {
		return err
	}

	// Verify location is stored as a sorted set
	zrangeTestCase := test_cases.ZrangeTestCase{
		Key:                 locationKey,
		StartIndex:          0,
		EndIndex:            -1,
		ExpectedMemberNames: []string{location.Name},
	}
	return zrangeTestCase.Run(client, logger)
}
