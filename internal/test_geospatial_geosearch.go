package internal

import (
	location_ds "github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialGeosearch(stageHarness *test_case_harness.TestCaseHarness) error {
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

	// Generate random locations
	locationSet := location_ds.GenerateRandomLocationSet(testerutils_random.RandomInt(4, 6))

	centerCoordinates := locationSet.CenterCoordinates()

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

	// Get 3 radii for testing:
	// 1. Smaller than the closest location
	// 2. Larger than farthest location
	// 3. Somewhere in between

	// GEOSEARCH with radius smaller than closest location from center - expect 0 results
	closestLocation := locationSet.ClosestTo(centerCoordinates)
	closestRadius := centerCoordinates.DistanceFrom(closestLocation.Coordinates)

	geosearchSmallRadiusTestCase := test_cases.GeoSearchTestCase{
		Key:                   locationKey,
		FromCoordinates:       centerCoordinates,
		Radius:                closestRadius * 0.75, // 3/4 of the smallest distance to include no locations
		ExpectedLocationNames: []string{},           // No locations expected
	}

	if err := geosearchSmallRadiusTestCase.Run(client, logger); err != nil {
		return err
	}

	// GEOSEARCH with radius larger than furthest location from center - expect all locations
	farthestLocation := locationSet.FarthestFrom(centerCoordinates)
	furthestRadius := centerCoordinates.DistanceFrom(farthestLocation.Coordinates)

	geosearchLargeRadiusTestCase := test_cases.GeoSearchTestCase{
		Key:                   locationKey,
		FromCoordinates:       centerCoordinates,
		Radius:                furthestRadius * 1.25,          // 1.25x greater than the furthest location to include all
		ExpectedLocationNames: locationSet.GetLocationNames(), // All locations expected
	}

	if err := geosearchLargeRadiusTestCase.Run(client, logger); err != nil {
		return err
	}

	// Test GEOSEARCH with radius in between - expect only the locations inside of the radius
	midRadius := (closestRadius + furthestRadius) / 2
	locationsInsideRadius := locationSet.WithinRadius(centerCoordinates, midRadius)

	geosearchMidRadiusTestCase := test_cases.GeoSearchTestCase{
		Key:                   locationKey,
		FromCoordinates:       centerCoordinates,
		Radius:                midRadius,
		ExpectedLocationNames: locationsInsideRadius.GetLocationNames(),
	}

	return geosearchMidRadiusTestCase.Run(client, logger)
}
