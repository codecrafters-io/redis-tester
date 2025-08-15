package internal

import (
	"sort"

	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
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

	locationKey := random.RandomWord()

	// Generate random locations
	locations := data_structures.GenerateRandomLocations(random.RandomInt(4, 6))

	referenceLocation := locations[0]
	locations = locations[1:]

	for _, location := range locations {
		geoAddTestCase := test_cases.NewGeoAddTestCaseWithValidCoordinates(locationKey, location, 1)
		if err := geoAddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	// Sort by increasing distance
	sort.Slice(locations, func(i, j int) bool {
		return referenceLocation.CalculateDistance(locations[i]) < referenceLocation.CalculateDistance(locations[j])
	})

	// Get 3 radii for testing:
	// 1. Smaller than the closest location
	// 2. Larger than farthest location
	// 3. Somewhere in between

	// GEOSEARCH with radius smaller than smallest distance - expect 0 results
	closestRadius := referenceLocation.CalculateDistance(locations[0])
	smallestRadius := closestRadius * 0.75 // 3/4 of the smallest distance to include no locations
	geosearchSmallRadiusTestCase := test_cases.GeoSearchTestCase{
		Key:                   locationKey,
		FromLongitude:         referenceLocation.GetLongitude(),
		FromLatitude:          referenceLocation.GetLatitude(),
		Radius:                smallestRadius,
		ExpectedLocationNames: []string{}, // No results expected
	}
	if err := geosearchSmallRadiusTestCase.Run(client, logger); err != nil {
		return err
	}

	// GEOSEARCH with radius larger than largest distance - expect all locations
	furthestRadius := referenceLocation.CalculateDistance(locations[len(locations)-1])
	largeRadius := furthestRadius * 1.25 // 1.25x greater than the furthest location to include all

	expectedLocationNames := make([]string, len(locations))
	for i, location := range locations {
		expectedLocationNames[i] = location.Name
	}
	geosearchLargeRadiusTestCase := test_cases.GeoSearchTestCase{
		Key:                   locationKey,
		FromLongitude:         referenceLocation.GetLongitude(),
		FromLatitude:          referenceLocation.GetLatitude(),
		Radius:                largeRadius,
		ExpectedLocationNames: expectedLocationNames,
	}
	if err := geosearchLargeRadiusTestCase.Run(client, logger); err != nil {
		return err
	}

	// Test GEOSEARCH with radius in between - expect only the locations inside of the radius
	midRadius := (closestRadius + furthestRadius) / 2
	expectedLocationNames = nil
	for _, location := range locations {
		if referenceLocation.CalculateDistance(location) < midRadius {
			expectedLocationNames = append(expectedLocationNames, location.Name)
		}
	}
	geosearchMidRadiusTestCase := test_cases.GeoSearchTestCase{
		Key:                   locationKey,
		FromLongitude:         referenceLocation.GetLongitude(),
		FromLatitude:          referenceLocation.GetLatitude(),
		Radius:                midRadius,
		ExpectedLocationNames: expectedLocationNames,
	}
	return geosearchMidRadiusTestCase.Run(client, logger)
}
