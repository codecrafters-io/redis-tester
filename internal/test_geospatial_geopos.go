package internal

import (
	"fmt"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialGeopos(stageHarness *test_case_harness.TestCaseHarness) error {
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

	locationKey := random.RandomWord()

	// Add locations
	locationSet := location.GenerateRandomLocationSet(random.RandomInt(2, 4))

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

	geoPosTestCase := test_cases.NewGeoPosTestCase(locationKey, false)
	geoPosTestCase.AddLocations(locationSet.GetLocations())

	missingLocations := make([]string, random.RandomInt(2, 4))
	for i := range len(missingLocations) {
		missingLocations[i] = fmt.Sprintf("missing_location_%d", random.RandomInt(1, 100))
	}

	geoPosTestCase.AddMissingLocations(missingLocations)
	if err := geoPosTestCase.Run(client, logger); err != nil {
		return err
	}

	// Test for missing key
	missingKey := fmt.Sprintf("missing_key_%d", random.RandomInt(1, 100))
	missingKeyTestCase := test_cases.NewGeoPosTestCase(missingKey, false)
	missingKeyTestCase.AddMissingLocations(missingLocations)
	return missingKeyTestCase.Run(client, logger)
}
