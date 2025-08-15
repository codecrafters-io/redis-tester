package internal

import (
	"fmt"
	"math"

	"github.com/codecrafters-io/redis-tester/internal/data_structures"
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
	locations := data_structures.GenerateRandomLocations(random.RandomInt(2, 4))
	for _, location := range locations {
		geoAddTestCase := test_cases.NewGeoAddTestCaseWithValidCoordinates(locationKey, location, 1)
		if err := geoAddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	locationNames := make([]string, len(locations))
	coordinates := make([]*data_structures.Coordinates, len(locations))
	for i, location := range locations {
		coordinates[i] = location.Coordinates
		locationNames[i] = location.Name
	}

	// Valid coordinates
	geoPosTestCase := test_cases.GeoPosTestCase{
		Key:                 locationKey,
		LocationNames:       locationNames,
		Tolerance:           math.Inf(1),
		ExpectedCoordinates: coordinates,
	}
	if err := geoPosTestCase.Run(client, logger); err != nil {
		return err
	}

	// Missing location
	missingLocationTestCase := test_cases.GeoPosTestCase{
		Key:                 locationKey,
		LocationNames:       []string{fmt.Sprintf("missing_key_%d", random.RandomInt(1, 100))},
		ExpectedCoordinates: []*data_structures.Coordinates{nil},
	}
	if err := missingLocationTestCase.Run(client, logger); err != nil {
		return err
	}

	// No locations
	noLocationsTestCase := test_cases.GeoPosTestCase{
		Key:                 locationKey,
		LocationNames:       nil,
		ExpectedCoordinates: nil,
	}
	if err := noLocationsTestCase.Run(client, logger); err != nil {
		return err
	}

	// Missing key - No locations
	missingKeyWithNoLocationsTestCase := test_cases.GeoPosTestCase{
		Key:                 fmt.Sprintf("missing_key_%d", random.RandomInt(1, 100)),
		LocationNames:       nil,
		ExpectedCoordinates: nil,
	}
	if err := missingKeyWithNoLocationsTestCase.Run(client, logger); err != nil {
		return err
	}

	// Missing key with a location
	missingKeyWithLocationTestCase := test_cases.GeoPosTestCase{
		Key:                 fmt.Sprintf("missing_key_%d", random.RandomInt(1, 100)),
		LocationNames:       []string{fmt.Sprintf("missing_location_%d", random.RandomInt(1, 100))},
		ExpectedCoordinates: []*data_structures.Coordinates{nil},
	}

	return missingKeyWithLocationTestCase.Run(client, logger)
}
