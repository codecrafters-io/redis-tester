package internal

import (
	location_ds "github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialGeodist(stageHarness *test_case_harness.TestCaseHarness) error {
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

	locationSet := location_ds.GenerateRandomLocationSet(testerutils_random.RandomInt(3, 5))
	locations := locationSet.GetLocations()

	for _, location := range locations {
		geoAddTestCase := test_cases.GeoAddTestCase{
			Key:                         locationKey,
			Location:                    location,
			ExpectedAddedLocationsCount: 1,
		}

		if err := geoAddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// Test distance calculations between all pairs of locations (6 test cases at max)
	for i := range locations {
		for j := i + 1; j < len(locations); j++ {
			geodistTestCase := test_cases.GeoDistTestCase{
				Key:              locationKey,
				Location1:        locations[i],
				Location2:        locations[j],
				ExpectedDistance: locations[i].DistanceFrom(locations[j]),
			}

			if err := geodistTestCase.Run(client, logger); err != nil {
				return err
			}
		}
	}

	// Test distance from a location to itself (should be 0)
	location := testerutils_random.RandomElementFromArray(locations)

	selfDistanceTestCase := test_cases.GeoDistTestCase{
		Key:              locationKey,
		Location1:        location,
		Location2:        location,
		ExpectedDistance: 0.0,
	}

	if err := selfDistanceTestCase.Run(client, logger); err != nil {
		return err
	}

	return nil
}
