package internal

import (
	"fmt"

	location_ds "github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
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

	locationKey := testerutils_random.RandomWord()

	// Add locations
	locationSet := location_ds.GenerateRandomLocationSet(testerutils_random.RandomInt(2, 4))

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

	missingLocationNames := make([]string, testerutils_random.RandomInt(2, 4))
	missing_location_suffix := testerutils_random.RandomInts(1, 100, len(missingLocationNames))

	for i := range len(missingLocationNames) {
		missingLocationNames[i] = fmt.Sprintf("missing_location_%d", missing_location_suffix[i])
	}

	geoPosTestCase := test_cases.GeoPosTestCase{
		Key:                              locationKey,
		Locations:                        locationSet.GetLocations(),
		MissingLocationNames:             missingLocationNames,
		ShouldSkipCoordinatesVerfication: true,
	}

	if err := geoPosTestCase.Run(client, logger); err != nil {
		return err
	}

	// Test for missing key
	missingKey := fmt.Sprintf("missing_key_%d", testerutils_random.RandomInt(1, 100))

	missingKeyTestCase := test_cases.GeoPosTestCase{
		Key:                              missingKey,
		MissingLocationNames:             missingLocationNames,
		ShouldSkipCoordinatesVerfication: true,
	}

	return missingKeyTestCase.Run(client, logger)
}
