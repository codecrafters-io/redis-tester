package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialCalculateScore(stageHarness *test_case_harness.TestCaseHarness) error {
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
	locationSet := location.GenerateRandomLocationSet(testerutils_random.RandomInt(2, 4))
	locations := locationSet.GetLocations()

	// Add locations
	for _, loc := range locations {
		geoAddTestCase := test_cases.NewGeoAddTestCaseWithAddedLocation(locationKey, loc, 1)
		if err := geoAddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// Check the score of each location
	logger.Infof("Checking the scores of added locations")

	for _, loc := range locations {
		zscoreTestCase := test_cases.SendCommandTestCase{
			Command:   "ZSCORE",
			Args:      []string{locationKey, loc.Name},
			Assertion: resp_assertions.NewFloatingPointBulkStringAssertion(float64(loc.GetGeoCode()), 0),
		}
		if err := zscoreTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	return nil
}
