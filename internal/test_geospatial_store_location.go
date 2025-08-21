package internal

import (
	location_ds "github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
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
	locationSet := location_ds.GenerateRandomLocationSet(testerutils_random.RandomInt(2, 4))

	// Add a location
	for _, location := range locationSet.GetLocations() {
		geoAddTestCase := test_cases.GeoAddTestCase{
			Key:                         locationKey,
			Location:                    location,
			ExpectedAddedLocationsCount: 1,
		}

		if err := geoAddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	logger.Infof("Checking if %q is stored as a sorted set", locationKey)
	// Verify location is stored as a sorted set
	// We don't use ZrangeTestCase as the score calculation stage is not covered yet, the order should not matter
	zrangeTestCase := test_cases.SendCommandTestCase{
		Command:   "ZRANGE",
		Args:      []string{locationKey, "0", "-1"},
		Assertion: resp_assertions.NewUnorderedStringArrayAssertion(locationSet.GetLocationNames()),
	}

	return zrangeTestCase.Run(client, logger)
}
