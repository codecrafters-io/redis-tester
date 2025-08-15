package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialDecodeCoordinates(stageHarness *test_case_harness.TestCaseHarness) error {
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

	locations := data_structures.GenerateRandomLocations(testerutils_random.RandomInt(2, 4))

	for _, location := range locations {
		zaddTestCase := test_cases.ZaddTestCase{
			Key: locationKey,
			Member: data_structures.SortedSetMember{
				Name:  location.Name,
				Score: float64(location.GetGeoCode()),
			},
			ExpectedAddedMembersCount: 1,
		}
		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	locationNames := make([]string, len(locations))
	expectedCoordinates := make([]*data_structures.Coordinates, len(locations))

	for i, location := range locations {
		locationNames[i] = location.Name
		currentCoordinate := location.GetGeoGridCenterCoordinates()
		expectedCoordinates[i] = &currentCoordinate
	}

	geoPosTestCase := test_cases.GeoPosTestCase{
		Key:                 locationKey,
		LocationNames:       locationNames,
		Tolerance:           10e-4,
		ExpectedCoordinates: expectedCoordinates,
	}

	return geoPosTestCase.Run(client, logger)
}
